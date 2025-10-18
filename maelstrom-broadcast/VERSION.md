# Version 1.0.b MultiNode Broadcast
In this version, we introduce a new feature that allows multiple nodes to broadcast messages to each other.
Each node stores a `currentVersion` counter. Whenever it ingests a brand new message—either through a direct broadcast or via gossip—the counter advances and that message is tagged with the fresh version.
Every second, the node will broadcast a message to one randomly selected node. For each of it's peers the node stores the last version of messages sent to that peer via gossip, that the peer acknolwedged
Only messages with a version strictly greater than the last recorded version (the latest value the peer has acknowledged) are sent to that peer.

## Benefits
- Efficient usage of bandwidth. We don't broadcast the entire message in each gossip, only the stuff we absolutely need to
- The version lock also makes sure we don't have to store too much state on any one node to keep track of what data the peer has received

## Caveats
- There is still some redundant data being sent during gossip, for example, two different nodes can send the same data to a third peer in gossip, because each node only tracks what it sent to a peer
- To avoid this, there would need to be some mechanism to track version data and state across all nodes which would significantly increase the complexity of the system
- Another caveat here is that there's an implicit assumption that broadcast messages are well distributed between all nodes. If a particular node doesn't receive fresh data for a while, its `currentVersion` keeps rising without new local broadcasts, and a single version can grow large—meaning gossip from this node may still carry a heavy batch when finally triggered.
