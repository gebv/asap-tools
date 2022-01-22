package version

import "runtime"

var (
	Version   string = "n/a"
	GitCommit string = "n/a"
	BuildDate string = "n/a"
	Goos             = runtime.GOOS
	Goarch           = runtime.GOARCH
)
