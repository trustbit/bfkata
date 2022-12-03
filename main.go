package main

import (
	"fmt"
	"github.com/trustbit/bfkata/cmd"
	"os"
)

var (
	Version   string
	GitCommit string
)

func main() {

	if len(os.Args) == 1 {
		cmd.PrintUsage()
		return
	}

	if len(Version) == 0 {
		fmt.Printf("Development version\n")
	} else {
		fmt.Printf("Version: %s Commit: %s\n", Version, GitCommit)
	}

	var code int
	switch os.Args[1] {
	case "test":
		fmt.Printf("")
		code = cmd.RunTests(os.Args[2:])
	case "api":
		code = cmd.PrintApi()
	case "specs":
		code = cmd.PrintSpecs(os.Args[2:])
	default:
		fmt.Printf("Unknown command %s\n", os.Args[1])
		cmd.PrintUsage()
		code = 1
	}

	os.Exit(code)

}
