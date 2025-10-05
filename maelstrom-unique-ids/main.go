package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func range_init() (int, int) {
	filePath := "./counter.txt"

	// Open or create the file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Repeatedly attempt to acquire lock (non-blocking)
	for {
		err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)

	// Read current counter value
	content, err := os.ReadFile(filePath)
	var currentCounter int
	if err != nil || len(content) == 0 {
		currentCounter = 0
	} else {
		currentCounter, err = strconv.Atoi(strings.TrimSpace(string(content)))
		if err != nil {
			currentCounter = 0
		}
	}

	// Write new counter value (original + 1001)
	newCounter := currentCounter + 10000
	err = file.Truncate(0)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.WriteString(strconv.Itoa(newCounter))
	if err != nil {
		log.Fatal(err)
	}

	return currentCounter, currentCounter + 9999
}

/**
 * Pretty simple logic. The file here acts as a counter,
 * in a production system this would be replicated by maybe a persistent redis node
 * The file keeps track of range of ids that has been used.
 * Each node that comes online will first have to get a range of ids, via the range_init function
 * Once a range is assigned, it's recorded in the file itself, by updating the stored value.
 * This ensures that each node get's a unique range and if it runs out, it requests a new range.
 * In case a node drops out before exhausting it's range, some number of IDs are lost
 */

/**
 * LIMITATION: IRL the supposed redis node or whatever, becomes a single point of failure.
 * If that node is down or inaccessible due to network issues, the system grinds to a halt.
 * One possible solution, is to have multiple redis nodes, each of which has a set number of ranges is can assign.
 * Whenever a redis node is about to run out of ranges, it will need to be refreshed
 * which should be a rare event and more tolerant to network issues or other transient failures
 */

/**
 * OPTIMIZATION: Can use multi-threading/go-routines here. Use a thread safe queue or counter to keep track of the counter/ID
 * or perhaps, assign the ID in main and then hand over the request and response ID to the goroutine, to complete?
 */
func main() {
	curr, end := range_init()

	n := maelstrom.NewNode()
	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = curr
		curr++

		err := n.Reply(msg, body)

		if curr >= end {
			curr, end = range_init()
		}
		return err
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
