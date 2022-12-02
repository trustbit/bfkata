package cmd

import "fmt"

func PrintUsage() {
	fmt.Printf(`
bfkata - test scaffolding for Black Friday kata. Commands:

  bfkata api       - print bundled contracts
  bfkata specs     - print bundled test specs
  bfkata test      - run test suite aginst a provided gRPC endpoint
`)
}
