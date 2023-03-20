package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"sync"
	"time"
)

var (
	counter      = Counter{value: 0}
	globalValues []int
	nodeMap      = make(map[string][]string) //map to store the topology
	globalSets   = make(map[int]int)
	readLock     sync.RWMutex
)

type ReadResp struct {
	MsgType  string `json:"type"`
	Messages []int  `json:"messages"`
}
type Task3aResp struct {
	MsgId   int    `json:"msg_id"`
	MsgType string `json:"type"`
}
type TopologyReq struct {
	Req
	Topology map[string][]string
}
type BroadcastReq struct {
	Req
	TrackKey string `json:"track_key"`
}
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

func (c *Counter) getNewCounter() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += 1

	return c.value
}

//main code for distributed workshop series with Shawn Nguyen
func distributedSession() {
	n := maelstrom.NewNode()

	//===========task 1===========
	n.Handle("echo", echoHandler(n))

	//===========task 2===========
	n.Handle("generate", generate1Handler(n))

	n.Handle("generate_cheat", generate2Handler(n))

	//===========task 3===========

	n.Handle("broadcast", broadcastHandlerTypeTotal(n))
	//n.Handle("broadcast", broadcastHandler(n))

	n.Handle("read", readHandler(n))

	n.Handle("topology", topologyHandler(nodeMap, n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func echoHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Update the message type to return back.
		body["type"] = "echo_ok"

		// Echo the original message back with the updated message type.
		return n.Reply(msg, body)
	}
}

func generate1Handler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
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
	}
}

func generate2Handler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body Req
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, Resp{
			MsgId:   body.MsgId,
			ID:      fmt.Sprintf("%d%d", body.MsgId, time.Now().UnixMicro()),
			MsgType: "generate_ok",
		})
	}
}

func broadcastHandlerTypeTotal(n *maelstrom.Node) func(msg maelstrom.
	Message) error {
	return func(msg maelstrom.Message) error {
		var body BroadcastReq
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		go n.Reply(msg, &Task3aResp{MsgId: body.MsgId,
			MsgType: "broadcast_ok"})

		//receive msg from client
		if len(body.TrackKey) == 0 {
			readLock.Lock()
			globalSets[body.Message] = 1
			readLock.Unlock()

			body.TrackKey = fmt.Sprintf("%s", msg.Dest)

			//broadcast to others
			if nodeMap[msg.Dest] != nil {
				for _, dest := range nodeMap[msg.Dest] {
					//n.Send(dest, body)
					go repeatSend(n, dest, body)
				}
			}
		} else { //receive msg from another node
			//check key exists
			readLock.RLock()
			_, ok := globalSets[body.Message]
			readLock.RUnlock()

			if ok {
				//do nothing
			} else {
				//add new value to global set
				readLock.Lock()
				globalSets[body.Message] = 1
				readLock.Unlock()
			}
		}
		return nil
	}
}

func broadcastHandler(n *maelstrom.Node) func(msg maelstrom.
	Message) error {
	return func(msg maelstrom.Message) error {
		var body BroadcastReq
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		n.Reply(msg, &Task3aResp{MsgId: body.MsgId,
			MsgType: "broadcast_ok"})

		//receive msg from client
		if len(body.TrackKey) == 0 {
			readLock.Lock()
			//globalValues = append(globalValues, body.Message)
			globalSets[body.Message] = 1
			readLock.Unlock()

			body.TrackKey = fmt.Sprintf("%s", msg.Dest)

			//broadcast to others
			if nodeMap[msg.Dest] != nil {
				neighbors := nodeMap[msg.Dest]
				for _, dest := range neighbors {
					//n.Send(dest, body)
					go repeatSend(n, dest, body)
				}
			}
		} else { //receive msg from another node
			//check key exists
			readLock.RLock()
			_, ok := globalSets[body.Message]
			readLock.RUnlock()

			if ok {
				//do nothing
			} else {
				//add new value to global set
				//log.Printf("before %v", globalValues)

				readLock.Lock()
				//globalValues = append(globalValues, body.Message)
				globalSets[body.Message] = 1
				readLock.Unlock()

				//log.Printf("after %v", globalValues)

				//broadcast to others
				neighbors := nodeMap[msg.Dest]
				for _, dest := range neighbors {
					if dest != msg.Src {
						//n.Send(dest, body)

						go repeatSend(n, dest, body)
					}
				}
			}
		}

		return nil
	}
}

func readHandler(n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		readLock.RLock()
		//values := globalValues

		values := []int{}

		for key := range globalSets {
			values = append(values, key)
		}
		readLock.RUnlock()

		return n.Reply(msg, ReadResp{MsgType: "read_ok",
			Messages: values})
	}
}

func topologyHandler(nodeMap map[string][]string, n *maelstrom.Node) func(msg maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body TopologyReq
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		go n.Reply(msg, &Task3aResp{MsgId: body.MsgId, MsgType: "topology_ok"})

		v := body.Topology[msg.Dest]
		clone := make([]string, len(v))
		copy(clone, v)
		key := msg.Dest
		nodeMap[key] = clone

		return nil
	}
}

var (
	DelayMillisecond time.Duration = 1000
)

func repeatSend(n *maelstrom.Node, dest string, body BroadcastReq) {
	clone := body
	stopSignal := make(chan int)

	//send msg to destination node
	go n.RPC(dest, clone, func(msg maelstrom.Message) error {
		stopSignal <- 1
		return nil
	})

	select {
	case <-stopSignal:
		return
	case <-time.After(DelayMillisecond * time.Millisecond):
		log.Printf("retry to send %+v", body)
		repeatSend(n, dest, body) //send again in case of no response
	}

}

func repeatSendNoAck(n *maelstrom.Node, dest string, body BroadcastReq) {
	clone := body
	//spam msg to other node
	for i := 1; i < 10; i++ {
		go n.Send(dest, clone)
		time.Sleep(DelayMillisecond)
	}
}

//local playground sections for try things
func testingFunc() {
	m := map[string]string{
		"java": "coffee",
		"go":   "verb",
		"ruby": "gemstone",
	}

	fmt.Println("len: ", len(m))

	val := make(chan int)
	go func() {
		time.Sleep(2 * time.Second)
		val <- 2
	}()

	select {
	case <-val:
		fmt.Println("ok")
		break
	case <-time.After(3 * time.Second):
		fmt.Println("not ok")
		break
	}
}

func main() {
	distributedSession()
	//testingFunc()
}
