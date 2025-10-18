package main

func newNodeState(nodeID string) *NodeState {
	return &NodeState{
		nodeID:         nodeID,
		currentVersion: 0,
		messages:       make(map[int]int),
		peerVersions:   make(map[string]int),
	}
}

// function used to add an incoming message via the broadcast API
// increments the currentVersion number marker
func (s *NodeState) addLocalMessage(value int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currentVersion++
	s.messages[value] = s.currentVersion
}

// function used to add incoming messages via the gossip API
// Does not increment currentVersion number marker
func (s *NodeState) addMessages(values []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	newDataSeen := false
	for _, value := range values {
		if _, ok := s.messages[value]; !ok { // if message is not already recorded in this node's state add it with the currentVersion number marker
			if !newDataSeen {
				newDataSeen = true
				s.currentVersion++
			}
			s.messages[value] = s.currentVersion
		}
	}
}

// Calculates delta for a given peer of this node
// This is based on the last known version of data that this peer acknowledged
// All messages with version equal to or greater than peers version are included in the delta
func (s *NodeState) deltaForPeer(peer string) ([]int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peerVersion := s.getPeerVersion(peer)
	out := make([]int, 0)
	maxVersion := peerVersion
	for value, version := range s.messages {
		if version > peerVersion {
			out = append(out, value)
		}
		maxVersion = max(maxVersion, version)
	}
	return out, maxVersion
}

func (s *NodeState) markPeerVersions(peer string, version int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentPeerVersion := s.getPeerVersion(peer)
	s.peerVersions[peer] = max(currentPeerVersion, version)
}

func (s *NodeState) snapshotMessages() []int {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]int, 0)
	for value := range s.messages {
		out = append(out, value)
	}
	return out
}

func (s *NodeState) getPeerVersion(peer string) int {
	version, ok := s.peerVersions[peer]
	if !ok {
		version = 0
		s.peerVersions[peer] = version
	}
	return version
}
