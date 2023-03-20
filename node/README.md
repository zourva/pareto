# s1 interface
1.generate protobuf files
```
protoc --go_out=plugins=grpc:. s1.proto
```

# TODO
- [ ] cluster mode support
  - [ ] raft based cluster
  - [ ] sqlite-based persistent