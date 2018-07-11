package version

import (
	"fmt"
	"os"
)

const ApplicationName string = "gitlab-slack-multiplexer"

func PrintAndExit() {
	info := getInfo()
	fmt.Printf("%s %s\n", ApplicationName, info.Version)
	fmt.Println(info)
	os.Exit(0)
}
