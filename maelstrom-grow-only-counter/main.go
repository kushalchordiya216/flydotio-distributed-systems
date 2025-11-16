package main

import (
	"context"
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	kv := maelstrom.NewLinKV(n)

	// Handler for /add endpoint
	n.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract delta from request
		delta, ok := body["delta"].(float64)
		if !ok {
			return nil
		}
		deltaInt := int(delta)

		ctx := context.Background()

		// CAS retry loop
		for {
			// Read current counter value
			current, err := kv.ReadInt(ctx, "counter")
			if err != nil {
				// Check if key doesn't exist
				if rpcErr, ok := err.(*maelstrom.RPCError); ok && rpcErr.Code == maelstrom.KeyDoesNotExist {
					current = 0
				} else {
					return err
				}
			}

			// Calculate new value
			newValue := current + deltaInt

			// Try to CAS
			err = kv.CompareAndSwap(ctx, "counter", current, newValue, true)
			if err == nil {
				// Success! Break out of retry loop
				break
			}

			// Check if it's a precondition failure (another update happened)
			if rpcErr, ok := err.(*maelstrom.RPCError); ok && rpcErr.Code == maelstrom.PreconditionFailed {
				// Retry - another node/handler updated the counter
				continue
			}

			// Some other error occurred
			return err
		}

		return n.Reply(msg, map[string]any{"type": "add_ok"})
	})

	// Handler for /read endpoint
	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		ctx := context.Background()

		// Read current counter value
		value, err := kv.ReadInt(ctx, "counter")
		if err != nil {
			// Check if key doesn't exist
			if rpcErr, ok := err.(*maelstrom.RPCError); ok && rpcErr.Code == maelstrom.KeyDoesNotExist {
				value = 0
			} else {
				return err
			}
		}

		body["type"] = "read_ok"
		body["value"] = value

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
