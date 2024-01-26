# Node Management
A node can be seen as a physical device running an operating system.

A node agent registers itself to a node server, and manages all services running on that node.

A node server manages all agents and services running on all nodes.

A service may not necessarily be managed to an agent, however, 
it always contains a messager to allow itself register quickly whenever it wants.
That means `messaging is mandatory while registration is on-demand`.

# s1 interface
The interface between a node server and an agent.

1.generate protobuf files
```
protoc --go_out=plugins=grpc:. s1.proto
```

# TODO
- [ ] cluster mode support
  - [ ] raft based cluster
  - [ ] sqlite-based persistent