# IPC
Implements both the BUS and RPC messaging patterns 
for both inter-process and inner-process peers.

Features:
- [x] BUS
  - [x] Publisher based on Nats MQ
  - [x] Subscriber based on Nats MQ
  - [x] Publisher based on inner-process EventBus
  - [x] Subscriber based on inner-process EventBus
- [x] RPC
  - [x] RPC Server based on Nats MQ
  - [x] RPC Client based on Nats MQ
  - [x] RPC Server based on inner-process moderator
  - [x] RPC Client based on inner-process moderator
  
Examples:
```
package main

import (
	"fmt"
	"github.com/zourva/pareto/ipc"
	"time"
)

func main() {
	// TODO
}

```
