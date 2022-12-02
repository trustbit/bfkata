package cmd

import "fmt"

func PrintUsage(version, commit string) {
	fmt.Printf(`
bfkata - test scaffolding for Black Friday kata. 

Version: %s, commit: %s

Commands:

  bfkata api       - print bundled contracts
  bfkata specs     - print bundled test specs
  bfkata test      - run test suite aginst a provided gRPC endpoint
`, version, commit)
}
