# Distributed system communication with Gossip Protocol (implemented with Golang)

## Prerequisites:
- [Maelstrom](https://github.com/jepsen-io/maelstrom/blob/main/doc/01-getting-ready/index.md): add maelstrom.jar file to lib folder

## Challenges:

### Echo:
```bash
./maelstrom test -w echo --bin ./main --node-count 1 --time-limit 10
```

### Unique Ids:
```bash
./maelstrom test -w unique-ids --bin ./main --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

### Broadcast:
#### a: Single-Node Broadcast
```bash
./maelstrom test -w broadcast --bin ./main --node-count 1 --time-limit 20 --rate 10
```
#### b: Multi-Node Broadcast
```bash
./maelstrom test -w broadcast --bin ./main --node-count 5 --time-limit 20 --rate 10
```
#### c: Fault Tolerant Broadcast
```bash
./maelstrom test -w broadcast --bin ./main --node-count 5 --time-limit 20 --rate 10 --nemesis partition
```
#### d & e: Efficient Broadcast
```bash
./maelstrom test -w broadcast --bin ./main --node-count 25 --time-limit 20 --rate 100 --topology total
```

```
:all {:send-count 42266,
             :recv-count 42266,
             :msg-count 42266,
             :msgs-per-op 25.354528},
       :clients {:send-count 3434, :recv-count 3434, :msg-count 3434},
       :servers {:send-count 38832,
                 :recv-count 38832,
                 :msg-count 38832,
                 :msgs-per-op 23.29454},
:stable-latencies {0 0, 0.5 80, 0.95 98, 0.99 198, 1 396},
```