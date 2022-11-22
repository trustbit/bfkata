package specs

import (
	"fmt"
	"github.com/abdullin/go-seq"
	"github.com/trustbit/bfkata/api"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"strings"
)

func yellow(s string) string {

	return fmt.Sprintf("%s%s%s", YELLOW, s, CLEAR)
}

const (
	CLEAR  = "\033[0m"
	RED    = "\033[91m"
	YELLOW = "\033[93m"

	GREEN = "\033[32m"

	ANOTHER = "\033[34m"
)

func red(s string) string {
	return fmt.Sprintf("%s%s%s", RED, s, CLEAR)
}

func PrintFull(s *api.Spec, issues seq.Issues) {

	if len(issues) > 0 {

		fmt.Printf(red("X %s (#%d)\n"), red(s.Name), s.Seq)
	} else {
		fmt.Printf("%sV %s %s(#%d)\n", GREEN, s.Name, CLEAR, s.Seq)
	}

	Print(s)
	/*
		if err != nil {
			fmt.Printf(red("  FATAL: %s\n"), err.Error())
		}
	*/
	if len(issues) == 0 {
		//fmt.Printf("%sNO ISSUES!%s\n", GREEN, CLEAR)

	} else {

		fmt.Println(yellow("ISSUES:"))

		for _, d := range issues {
			fmt.Printf("  %sΔ %s%s\n", ANOTHER, IssueToString(d), CLEAR)
		}
	}

}
func IssueToString(d seq.Issue) string {
	return fmt.Sprintf("Expected %v to be %v but got %v",
		strings.Replace(seq.JoinPath(d.Path), ".[", "[", -1),
		Format(d.Expected),
		Format(d.Actual))

}

func Print(s *api.Spec) {
	//println(s.Name)

	if len(s.Given) > 0 {
		println(yellow("GIVEN:"))
		for i, e := range s.Given {
			println(fmt.Sprintf("  %d. %s", i+1, Format(e)))
		}
	}
	println(fmt.Sprintf("%s\n  %s", yellow("WHEN:"), Format(s.When)))
	if s.ThenResponse != nil {
		println(fmt.Sprintf("%s\n  %s", yellow("THEN RESPONSE:"), Format(s.ThenResponse)))
	}
	if s.ThenError != nil {
		println(fmt.Sprintf("%s\n  %s", yellow("THEN ERROR:"), statusToString(s.ThenError)))
	}
	if len(s.ThenEvents) > 0 {
		println(yellow("THEN EVENTS:"))
		for i, e := range s.ThenEvents {
			println(fmt.Sprintf("  %d. %s", i+1, Format(e)))
		}
	}
}
func Format(val any) string {

	if val == nil {
		return "<nil>"
	}
	switch v := val.(type) {

	case proto.Message:

		repr := prototext.MarshalOptions{Multiline: false}.Format(v)
		return string(v.ProtoReflect().Descriptor().Name()) + " " + repr + ""
	case []proto.Message:
		names := []string{}
		for _, m := range v {
			names = append(names, string(m.ProtoReflect().Descriptor().Name()))
		}
		return fmt.Sprintf("[%s]", strings.Join(names, ", "))

	case error:

		st, ok := status.FromError(v)
		if ok {
			return fmt.Sprintf("\"%s\" (%v)", st.Message(), st.Code().String())
		} else {

			return fmt.Sprintf("Error '%v'", v.Error())
		}
	default:
		return fmt.Sprintf("'%v'", v)
	}
}
