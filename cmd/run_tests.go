package cmd

import (
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
)

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

func RunTests(args []string) int {
	var addr string
	var file string
	var specNum int
	var limit int
	flags := flag.NewFlagSet("test", flag.ExitOnError)

	flags.StringVar(&addr, "addr", "127.0.0.1:50051", "Subject to test")
	flags.StringVar(&file, "file", specs.BUNDLE, "Specs file to load")
	flags.IntVar(&specNum, "spec", 0, "Spec id to explore")
	flags.IntVar(&limit, "limit", 3, "Max failures to show (-1 to limit)")

	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return 1
	}

	actual, err := specs.Load(file)
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
		fmt.Printf("OK! Kata implemented by %s%s%s (%s)\n\n",
			YELLOW,
			resp.Author, CLEAR, resp.Detail)
	}

	for i, s := range actual {

		fmt.Printf("#%d. %s%s%s...", i+1, YELLOW, s.Name, CLEAR)

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

			if limit != -1 {
				if fails < limit {
					specs.PrintFull(s, deltas)
					println()
				}

			}

			fails += 1
		}

	}

	if fails > limit {
		fmt.Printf("%s%d failing spec(s) hidden.%s Use 'limit' flag to show more\n", RED, limit, CLEAR)
	}

	fmt.Printf("%sPass:%d%s %sFail:%d%s Deltas:%d\n",
		GREEN, oks, CLEAR,
		RED, fails, CLEAR, issues)

	if fails > 0 {
		return 1
	}
	return 0
}
