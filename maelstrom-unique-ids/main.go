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
