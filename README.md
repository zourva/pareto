# pareto
Frequently used go code for 80% projects when programming.

Features:
- [x] basics(box)
  - [x] type converters
  - [x] encoder & decoders
  - [x] file operations
  - [x] math functions
  - [x] network operations
  - [x] timestamps
  - [x] cpu id
- [ ] services
  - [x] logging
  - [x] profiling
  - [x] environment management
  - [ ] resource management
    - [x] string literals 
- [ ] meta pattern:
  - [x] loop controller 
  - [x] state machine


Examples:
```
package main

import (
	"fmt"
	"github.com/zourva/pareto"
	"github.com/zourva/pareto/meta"
	"time"
)

func main() {
	pareto.Setup()

	loop := meta.NewLoop("monitor", meta.LoopConfig{
		Tick:        1000,  // tick interval, 1000 ms
		Work:        1,     //event callback trigger ticks
		Sync:        false, //execute callback in a separate goroutine
		BailOnError: false,
	})

	loop.Run(meta.LoopRunHook{
		Working: func() error {
			fmt.Println("monitoring...")
			return nil
		},
	})

	time.Sleep(time.Second * 10)

	loop.Stop()

	pareto.Teardown()
}

```
