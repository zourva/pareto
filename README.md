# pareto
Frequently used go code for 80% projects when programming.

NOT stable yet and may change dramatically.

Features:
- [ ] basics(box)
  - [x] type converters
  - [x] encoder & decoders
  - [x] file operations
  - [x] math functions
  - [x] network operations
  - [x] timestamps
  - [x] cpu id
- [ ] meta patterns
  - [x] loop controller 
  - [x] state machine
- [ ] network topology
  - [x] MQTT broker embedded based on [Mochi](https://github.com/mochi-co/mqtt/server)
  - [x] MQTT client based on [paho](https://github.com/eclipse/paho.golang/paho)
  - [x] MQ client based on [nats]()
  - [x] RPC server based on [msgpack-rpc](#)
  - [x] RPC client based on [msgpack-rpc](#)
  - [x] inter-module communication 
    - [x] inner-proc event bus based on [EventBus](https://github.com/asaskevich/EventBus)
- [ ] distributed node management
  - [ ] node provisioning & configuration
  - [ ] node status keeping & monitoring
  - [ ] node online/offline(join & leave)
  - [ ] admin api
    - [x] node info maintenance
    - [x] uploader & downloader
- [ ] inner-process services
  - [x] logging based on [logrus](https://github.com/sirupsen/logrus)
  - [x] profiling
  - [ ] config management based on [viper](https://github.com/spf13/viper)
  - [ ] options management based on [cobra](https://github.com/spf13/cobra)
  - [x] working directory management
  - [x] resource management
    - [x] string literals 

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
