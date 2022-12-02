package main

import (
	"fmt"
	"github.com/trustbit/bfkata/cmd"
	"os"
)

func main() {

	if len(os.Args) == 1 {
		cmd.PrintUsage()
		return
	}

	var code int
	switch os.Args[1] {
	case "test":
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
