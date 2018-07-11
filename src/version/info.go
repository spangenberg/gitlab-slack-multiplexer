package version

import (
	"fmt"
	"runtime"
)

var (
	version   string
	buildDate string
)

type info struct {
	Version   string
	BuildDate string
	GoVersion string
}

func (v info) String() string {
	return fmt.Sprintf("%#v", v)
}

func getInfo() info {
	return info{
		Version:   version,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
	}
}
