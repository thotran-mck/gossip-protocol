package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
	"time"
)

type Req struct {
	MsgId   int    `json:"msg_id"`
	MsgType string `json:"type"`
	Message int    `json:"message"`
}

type Resp struct {
	MsgId   int    `json:"msg_id"`
	ID      string `json:"id"`
	MsgType string `json:"type"`
}

type Counter struct {
	mutex sync.Mutex
	value int64
}

func (c *Counter) incCounter() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.value < 0 {
		c.value = 0
	}

	c.value += 1
}

func (c *Counter) getNewCounter() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += 1

	return c.value
}

var (
	counter = Counter{value: 0}
	//for cheating purpose
	lock     sync.Mutex
	globalId int64 = 1
)

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//main code for distributed workshop series with Shawn Nguyen
func distributedSession() {
	n := maelstrom.NewNode()

	//===========task 1===========
	n.Handle("echo", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "echo_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	})

	//===========task 2===========
	n.Handle("generate", func(msg maelstrom.Message) error {
		var body Req
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, Resp{
			MsgId: body.MsgId,
			ID: fmt.Sprintf("%d-%d", time.Now().UnixMicro(),
				counter.getNewCounter()),
			MsgType: "generate_ok",
		})
	})

	n.Handle("generate_cheat", func(msg maelstrom.Message) error {
		var body Req
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, Resp{
			MsgId:   body.MsgId,
			ID:      fmt.Sprintf("%d%d", body.MsgId, time.Now().UnixMicro()),
			MsgType: "generate_ok",
		})
	})

	//===========task 3===========
	var (
		globalSet []int
		nodeMap   = make(map[string][]string) //map to store the topology
		//historyMap = make(map[string]int)      //map to keep track update history
	)

	type ReadResp struct {
		MsgType  string `json:"type"`
		Messages []int  `json:"messages"`
	}
	type Task3aResp struct {
		MsgId   int    `json:"msg_id"`
		MsgType string `json:"type"`
	}

	type BroadcastReq struct {
		Req
		TrackKey string `json:"track_key"`
	}

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastReq
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		if len(body.TrackKey) == 0 {
			globalSet = append(globalSet, body.Message)

			//save to history map
			timeKey := fmt.Sprintf("%s_%d", msg.Dest, time.Now().UnixMicro())

			//lock.Lock()
			//historyMap[timeKey] = body.Message
			//lock.Unlock()

			body.TrackKey = timeKey

			//broadcast to others
			if nodeMap[msg.Dest] != nil {
				neighbors := nodeMap[msg.Dest]
				for _, dest := range neighbors {

					go n.Send(dest, body)
				}
			}
		} else {
			//check key exists
			ok := contains(globalSet, body.Message)

			if ok {
				//do nothing
			} else {
				//add new value to global set
				globalSet = append(globalSet, body.Message)

				//save history
				//lock.Lock()
				//historyMap[body.TrackKey] = body.Message
				//lock.Unlock()

				//broadcast to others
				neighbors := nodeMap[msg.Dest]
				for _, dest := range neighbors {
					go n.Send(dest, body)
				}
			}
		}

		return n.Reply(msg, &Task3aResp{MsgId: body.MsgId,
			MsgType: "broadcast_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body Req
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body.MsgType = "read_ok"
		resp := ReadResp{MsgType: "read_ok", Messages: globalSet}
		return n.Reply(msg, resp)
	})

	type TopologyReq struct {
		Req
		Topology map[string][]string
	}

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body TopologyReq
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		//nodeMap = body.Topology
		for k, v := range body.Topology {
			clone := make([]string, len(v))
			copy(clone, v)
			key := k
			nodeMap[key] = clone
		}

		return n.Reply(msg, &Task3aResp{MsgId: body.MsgId, MsgType: "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

//local playground sections for try things
func testingFunc() {
	//fmt.Println("Welcome to the playground! Here is your session number: ", getRandomInt())

	//counter := Counter{value: 0}
	//for i := 1; i < 10000000; i++ {
	//	go counter.getNewCounter()
	//}
	//time.Sleep(time.Second)
	//fmt.Print("final value: ", counter.value)

	var mp = make(map[string]int)
	_, ok := mp["12"]
	if ok {
		fmt.Println("YES")
	} else {
		fmt.Println("no")
	}
}

func main() {
	distributedSession()
	testingFunc()
}
