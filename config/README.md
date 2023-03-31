# Config
A configuration management library based on [BoltDB](https://go.etcd.io/bbolt) database.

Features:
- [x] Underlying database file management
- [x] Configuration writer APIs
    - [ ] SetBool(section string, k string, v bool)
    - [ ] SetFloat(section string, k string, v float64)
    - [ ] SetInt(section string, k string, v int)
    - [ ] SetString(section string, k string, v string)
    - [ ] Remove(section string, k string)
    - [ ] AddChangeListener(section string, func(k string, v interface{}))
- [x] Configuration reader APIs
    - [ ] Bool(k string) bool
    - [ ] BoolOpt(k string, fallback bool) bool
    - [ ] Float(k string) float64
    - [ ] FloatOpt(k string, fallback float64) float64
    - [ ] Int(k string) int
    - [ ] IntOpt(k string, fallback int) int
    - [ ] String(k string) string
    - [ ] StringOpt(k string, fallback string) string
  
Examples:
```
package main

import (
	"fmt"
	"github.com/zourva/pareto/config"
	"time"
)

func main() {
	// TODO
}

```
