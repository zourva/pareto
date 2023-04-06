package builder

// Injectable variables using go build -ldflags "
//   - -X 'builder.Organization=Org'
//   - -X 'builder.Product=Prd'
//   - -X 'builder.Version=`cat VERSION`'
//   - -X 'builder.GoVersion=`go env GOVERSION`'
//   - -X 'builder.GitCommit=`git rev-parse HEAD`'
//   - -X 'builder.OSArch=`go env GOOS`/`go env GOARCH`'
//   - -X 'builder.BuildTime=`date +%Y%m%d%H%M%S`'
var (
	Organization = "unknown"
	Product      = "unknown"
	Version      = "unknown"

	GoVersion = "unknown"
	GitCommit = "unknown"
	OSArch    = "unknown"
	BuildTime = "unknown"
)
