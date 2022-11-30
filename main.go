package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/trustbit/bfkata/api"
	"github.com/trustbit/bfkata/specs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"os"
	"regexp"
	"strings"
)

const BUNDLE = "<bundle>"

func main() {

	if len(os.Args) == 1 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "test":
		RunTest(os.Args[2:])
	case "api":
		printApi()
		return
	case "specs":
		printSpecs(os.Args[2:])
	default:
		fmt.Printf("Unknown command %s", os.Args[1])
		printUsage()
		return

	}

}

var comment = regexp.MustCompile(`//.*`)

func printApi() {
	// poor man's keyword highlight
	keywords := []string{"message",
		"service", "returns", "rpc", "string", "repeated", "int32", "enum", "int64", "map<string,string>", "google.protobuf.Any"}
	txt := api.BundledAPI
	for _, kw := range keywords {
		txt = strings.Replace(txt, kw+" ", GREEN+kw+CLEAR+" ", -1)
	}

	txt = comment.ReplaceAllStringFunc(txt, func(s string) string {
		return YELLOW + s + CLEAR
	})

	fmt.Println(txt)
}

func printSpecs(args []string) int {

	var compact bool
	flags := flag.NewFlagSet("specs", flag.ExitOnError)

	flags.BoolVar(&compact, "compact", false, "Display headers only")

	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return 1
	}

	sp, err := loadSpecs(BUNDLE)
	if err != nil {
		log.Fatalln(err)
	}

	if compact {
		for _, s := range sp {
			fmt.Printf("%2d %s\n", s.Seq, s.Name)
		}
	} else {

		fmt.Printf("// Loaded %d specs from %s\n", len(sp), BUNDLE)
		for _, s := range sp {
			fmt.Println(specs.BODY_SEPARATOR)
			fmt.Printf("%s%s %s(#%d)\n", GREEN, s.Name, CLEAR, s.Seq)
			fmt.Println(specs.NAME_SEPARATOR)
			specs.Print(s)

		}
	}

	return 0
}

func printUsage() {
	fmt.Printf(`
bfkata - test scaffolding for Black Friday kata. Commands:

  api       - print bundled contracts
  specs     - print bundled test specs
  test      - run test suite aginst a provided gRPC endpoint
`)
}

const (
	CLEAR  = "\033[0m"
	RED    = "\033[91m"
	YELLOW = "\033[93m"

	GREEN = "\033[32m"

	ANOTHER = "\033[34m"
	ERASE   = "\033[2K"
)

func red(s string) string {
	return fmt.Sprintf("%s%s%s", RED, s, CLEAR)
}
func yellow(s string) string {

	return fmt.Sprintf("%s%s%s", YELLOW, s, CLEAR)
}

func green(s string) string {

	return fmt.Sprintf("%s%s%s", GREEN, s, CLEAR)
}

func mustAny(p proto.Message) *anypb.Any {
	r, err := anypb.New(p)
	if err != nil {
		log.Panicln("failed to convert to any: %w", err)
	}
	return r
}

func mustMsg(a *anypb.Any) proto.Message {
	if a == nil {
		return nil
	}
	p, err := a.UnmarshalNew()
	if err != nil {
		log.Panicln("failed to convert from any: %w", err)
	}
	return p
}

func RunTest(args []string) int {

	fmt.Println(args)
	var addr string
	var file string
	var specNum int
	flags := flag.NewFlagSet("test", flag.ExitOnError)

	flags.StringVar(&addr, "addr", "127.0.0.1:50051", "Subject to test")
	flags.StringVar(&file, "file", BUNDLE, "Specs file to load")
	flags.IntVar(&specNum, "spec", 0, "Spec id to explore")

	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return 1
	}

	actual, err := loadSpecs(file)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	if specNum > 0 {
		actual = actual[specNum : specNum+1]
	}

	fmt.Printf("Loaded %d specs from %s\n", len(actual), file)

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println("Can't dial address:", err)
		return 1
	}
	// setup client
	client := api.NewSpecServiceClient(conn)

	// speed test

	oks, fails, issues := 0, 0, 0

	fmt.Printf("Connecting to %s...\n", addr)

	resp, err := client.About(ctx, &api.AboutRequest{})
	if err != nil {

		fail, _ := status.FromError(err)
		if fail.Code() == codes.Unavailable {

			fmt.Println(fail.Message())

			fmt.Printf("\nTest endpoint is not found. Did you start it?\n")
			return 1

		}
		fmt.Println(err.Error())
		return 1
	} else {
		fmt.Printf("OK! Implemented by %s\n", resp.Author)
	}

	for i, s := range actual {

		fmt.Printf("#%d. %s...", i+1, yellow(s.Name))

		request := &api.SpecRequest{
			When: mustAny(s.When),
		}

		for _, e := range s.Given {
			request.Given = append(request.Given, mustAny(e))
		}

		resp, err := client.Spec(ctx, request)

		if err != nil {
			log.Fatalln(err)
		}
		var events []proto.Message
		for _, e := range resp.Events {
			events = append(events, mustMsg(e))
		}

		st := status.New(codes.Code(resp.Status), resp.Error)

		deltas := specs.Compare(s, mustMsg(resp.Response), st, events)
		issues += len(deltas)

		fmt.Print(ERASE, "\r")
		if len(deltas) == 0 && err == nil {
			oks += 1
		} else {
			fails += 1
			specs.PrintFull(s, deltas)
			println()
		}

	}
	fmt.Printf("Pass:%d Fail:%d Deltas:%d\n", oks, fails, issues)
	return 0
}

func loadSpecs(file string) ([]*api.Spec, error) {
	var reader *bytes.Reader
	if file == BUNDLE {
		reader = bytes.NewReader([]byte(specs.BundledSpecs))

	} else {
		in, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("can't read file: %w", err)
		}
		reader = bytes.NewReader(in)
	}
	actual, err := specs.ReadSpecs(reader)
	if err != nil {
		return nil, fmt.Errorf("can't parse specs:", err)
	}
	return actual, nil

}
