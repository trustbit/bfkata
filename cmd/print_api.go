package cmd

import (
	"fmt"
	"github.com/trustbit/bfkata/api"
	"regexp"
	"strings"
)

var comment = regexp.MustCompile(`//.*`)

func PrintApi() int {
	// poor man's keyword highlight
	keywords := []string{
		"message",
		"service",
		"returns",
		"rpc",
		"string",
		"repeated",
		"int32",
		"enum",
		"int64",
		"map<string,string>",
		"google.protobuf.Any",
	}
	txt := api.BundledAPI
	for _, kw := range keywords {
		txt = strings.Replace(txt, kw+" ", GREEN+kw+CLEAR+" ", -1)
	}

	txt = comment.ReplaceAllStringFunc(txt, func(s string) string {
		return YELLOW + s + CLEAR
	})

	fmt.Println(txt)

	return 0
}
