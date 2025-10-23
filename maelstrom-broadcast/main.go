package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	state := newNodeState(n.ID())
	gossipInterval := 50 * time.Millisecond

	go func() {
		for {
			time.Sleep(gossipInterval)
			peers := peerList(n)
			if len(peers) == 0 {
				continue
			}

			target := peers[rand.Intn(len(peers))]
			delta, version := state.deltaForPeer(target)
			if len(delta) == 0 {
				continue
			}

			payload := GossipPayload{
				Type:     "gossip",
				Messages: delta,
			}

			n.RPC(target, payload, func(reply maelstrom.Message) error {
				state.markPeerVersions(target, version) // gossip acknowledged

				var body GossipPayload
				if err := json.Unmarshal(reply.Body, &body); err != nil {
					return err
				}
				state.addMessages(body.Messages)
				return nil
			})
		}
	}()

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body BroadcastPayload
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		state.addLocalMessage(body.Message)
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("gossip", func(msg maelstrom.Message) error {
		var body GossipPayload
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		state.addMessages(body.Messages)
		delta, version := state.deltaForPeer(msg.Src)
		reply := GossipPayload{
			Type:     "gossip_ok",
			Messages: delta,
		}
		// risque - network partition could fail halfway through, meaning we are marking something as being acknowledged before it is actually acknowledged
		// Ideally, we make another RPC so that we can get back an ack
		state.markPeerVersions(msg.Src, version)

		return n.Reply(msg, reply)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		response := map[string]any{
			"type":     "read_ok",
			"messages": state.snapshotMessages(),
		}
		return n.Reply(msg, response)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body TopologyPayload
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func peerList(n *maelstrom.Node) []string {
	raw := n.NodeIDs()
	out := make([]string, 0, len(raw))
	for _, id := range raw {
		if id != n.ID() {
			out = append(out, id)
		}
	}
	return out
}
