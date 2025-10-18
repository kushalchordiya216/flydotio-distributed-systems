package main

import "sync"

type BroadcastPayload struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type GossipPayload struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

type TopologyPayload struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

type NodeState struct {
	mu             sync.Mutex
	nodeID         string
	currentVersion int
	messages       map[int]int
	peerVersions   map[string]int
}
