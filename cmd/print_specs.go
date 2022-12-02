package cmd

import (
	"flag"
	"fmt"
	"github.com/trustbit/bfkata/specs"
	"log"
)

func PrintSpecs(args []string) int {

	var compact bool
	flags := flag.NewFlagSet("specs", flag.ExitOnError)

	flags.BoolVar(&compact, "compact", false, "Display headers only")

	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return 1
	}

	sp, err := specs.Load(specs.BUNDLE)
	if err != nil {
		log.Fatalln(err)
	}

	if compact {
		for _, s := range sp {
			fmt.Printf("%2d %s\n", s.Seq, s.Name)
		}
	} else {

		fmt.Printf("// Loaded %d specs from %s\n", len(sp), specs.BUNDLE)
		for _, s := range sp {
			fmt.Println(specs.BODY_SEPARATOR)
			fmt.Printf("%s%s %s(#%d)\n", GREEN, s.Name, CLEAR, s.Seq)
			fmt.Println(specs.NAME_SEPARATOR)
			specs.Print(s)

		}
	}

	return 0
}
