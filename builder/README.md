# Builder
This package abstracts the settings necessary for building of executables.

Injectable variables through `go build -ldflags -X` :
- [x] organization info
- [x] product info
- [x] version info
- [x] build environment
   - [x] go version
   - [x] os and arch
   - [x] git commit id
   - [x] build timestamp
  
Examples:
```
go build -ldflags -X 'builder.Version=1.0' -o main.exe main.go
```
