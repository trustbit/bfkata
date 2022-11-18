package specs

import (
	"fmt"
	"github.com/abdullin/go-seq"
	"github.com/trustbit/bfkata/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func errToStr(e *status.Status) string {
	if e == nil {
		return ""
	}
	return e.String()

}

func Compare(spec *api.Spec, resp proto.Message, actualErr *status.Status, events []proto.Message) seq.Issues {

	issues := seq.Diff(spec.ThenResponse, resp, "response")

	expectedStatus := codes.OK
	if spec.ThenError != nil {
		expectedStatus = spec.ThenError.Code()
	}
	actualStatus := codes.OK
	if actualErr != nil {
		actualStatus = actualErr.Code()
	}

	if expectedStatus != actualStatus {

		issues = append(issues, seq.Issue{
			Expected: expectedStatus,
			Actual:   actualStatus,
			Path:     []string{"status"},
		})

	}

	if len(events) != len(spec.ThenEvents) {
		issues = append(issues, seq.Issue{
			Expected: spec.ThenEvents,
			Actual:   events,
			Path:     []string{"Events"},
		})
	} else {
		for i, e := range spec.ThenEvents {
			p := []string{"Events", fmt.Sprintf("[%d]", i)}
			issues = append(issues, seq.Diff(e, events[i], p...)...)
		}
	}
	return issues
}
