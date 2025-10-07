package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	// Store for broadcast messages
	var messages []int
	messageChan := make(chan int, 100)

	// Store for topology
	var topology map[string][]string

	// Goroutine to consume from channel and add to messages store
	go func() {
		for msg := range messageChan {
			messages = append(messages, msg)
		}
	}()

	// Handle broadcast - receives a single message (int) and stores it
	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract the message value
		message := int(body["message"].(float64))

		// Send to channel
		messageChan <- message

		// Send acknowledgment
		response := make(map[string]any)
		response["type"] = "broadcast_ok"

		return n.Reply(msg, response)
	})

	// Handle read - returns all broadcasts so far
	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Copy messages for response
		messagesCopy := make([]int, len(messages))
		copy(messagesCopy, messages)

		// Send response with all messages
		response := make(map[string]any)
		response["type"] = "read_ok"
		response["messages"] = messagesCopy

		return n.Reply(msg, response)
	})

	// Handle topology - receives a topology of type map[string][]string
	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract and store topology
		if topo, ok := body["topology"].(map[string]any); ok {
			topology = make(map[string][]string)
			for key, value := range topo {
				if neighbors, ok := value.([]any); ok {
					strNeighbors := make([]string, len(neighbors))
					for i, neighbor := range neighbors {
						strNeighbors[i] = neighbor.(string)
					}
					topology[key] = strNeighbors
				}
			}
		}

		// Send acknowledgment
		response := make(map[string]any)
		response["type"] = "topology_ok"

		return n.Reply(msg, response)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
