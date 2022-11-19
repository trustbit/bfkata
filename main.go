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
)

func main() {
	runTest()
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

func runTest() int {
	var addr string
	var file string
	flags := flag.NewFlagSet("test", flag.ExitOnError)

	flags.StringVar(&addr, "addr", "127.0.0.1:50051", "Subject to test")
	flags.StringVar(&file, "file", "<bundled>", "Specs file to load")

	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return 1
	}
	var reader *bytes.Reader
	if file == "<bundled>" {
		reader = bytes.NewReader([]byte(specs.BundledSpecs))

	} else {
		in, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("Can't read file:", err)
			return 1
		}

		reader = bytes.NewReader(in)
	}
	actual, err := specs.ReadSpecs(reader)
	if err != nil {
		fmt.Println("Can't read specs:", err)
		return 1
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
