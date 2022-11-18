package specs

import (
	"bufio"
	"fmt"
	"github.com/trustbit/bfkata/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"io"
	"log"
	"strings"
)

func msgToString(m proto.Message) string {
	body := prototext.MarshalOptions{
		Multiline: false,
	}.Format(m)

	return fmt.Sprintf("%s %s", m.ProtoReflect().Descriptor().Name(), body)
}

func stringsToMsgs(lines []string) ([]proto.Message, error) {
	var msgs []proto.Message
	for _, l := range lines {
		m, err := stringToMsg(l)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func stringToMsg(s string) (proto.Message, error) {

	source := strings.TrimSpace(s)
	splits := strings.SplitN(source, " ", 2)

	name := splits[0]

	mt, found := lookups[name]
	if !found {
		return nil, fmt.Errorf("Unknown type %s in line %s", name, s)
	}

	instance := mt.New().Interface()

	if len(splits) == 2 {
		body := []byte(splits[1])

		if err := prototext.Unmarshal(body, instance); err != nil {
			return nil, err
		}
	}

	return instance, nil

}

const NAME_SEPARATOR = "------------------------------------------"
const BODY_SEPARATOR = "=========================================="

func WriteSpecs(specs []*api.Spec, w io.Writer) error {

	for i, s := range specs {
		if i > 0 {
			if _, err := fmt.Fprintln(w, BODY_SEPARATOR); err != nil {
				return err
			}
		}
		if str, err := SpecToParseableString(s); err != nil {
			return err
		} else {
			if _, err := fmt.Fprintln(w, str); err != nil {
				return err
			}
		}
	}
	return nil
}

func ReadSpecs(r io.Reader) ([]*api.Spec, error) {
	scanner := bufio.NewScanner(r)

	var specs []*api.Spec
	var lines []string
	var line int
	for scanner.Scan() {
		line += 1
		s := strings.TrimSpace(scanner.Text())
		if len(s) == 0 {
			continue
		}

		// minimal separator
		if strings.HasPrefix(s, "===") {
			if len(lines) > 0 {

				// hacky for now
				joined := strings.Join(lines, "\n")
				parsed, err := SpecFromParseableString(joined)
				if err != nil {
					return nil, fmt.Errorf("failed to parse spec from line %d: %w", line, err)
				}
				lines = nil
				specs = append(specs, parsed)
			}
		} else {
			lines = append(lines, s)

		}

	}

	if len(lines) > 0 {

		// hacky for now
		joined := strings.Join(lines, "\n")
		parsed, err := SpecFromParseableString(joined)
		if err != nil {
			return nil, fmt.Errorf("failed to parse spec from line %d: %w", line, err)
		}
		specs = append(specs, parsed)
	}
	return specs, nil
}

func SpecToParseableString(s *api.Spec) (string, error) {

	var b strings.Builder
	ln := func(text string, args ...any) {
		_, err := b.WriteString(fmt.Sprintf(text, args...) + "\n")
		if err != nil {
			panic(err)
		}
	}

	ln(s.Name)
	ln(NAME_SEPARATOR)
	ln("GIVEN:")

	for _, e := range s.Given {
		ln("  %s", msgToString(e))
	}
	ln("WHEN:\n  %s", msgToString(s.When))
	if s.ThenResponse != nil {
		ln("THEN:\n  %s", msgToString(s.ThenResponse))
	}

	if len(s.ThenEvents) > 0 {
		ln("EVENTS:")
		for _, e := range s.ThenEvents {
			ln("  %s", msgToString(e))
		}
	}
	if s.ThenError != nil {
		ln("ERROR:\n  %s", statusToString(s.ThenError))
	}

	return b.String(), nil

}

var lookups = make(map[string]protoreflect.MessageType)
var strToCode = make(map[string]codes.Code)

func init() {
	msgs := api.File_api_api_proto.Messages()

	for i := 0; i < msgs.Len(); i++ {
		m := msgs.Get(i)

		mt, err := protoregistry.GlobalTypes.FindMessageByName(m.FullName())
		if err != nil {
			log.Panicln(err)
		}

		lookups[string(m.Name())] = mt
	}

	for i := codes.OK; i <= codes.Unauthenticated; i++ {
		strToCode[i.String()] = i
	}
}

func groupLines(lines []string) (map[string][]string, error) {
	current := ""
	result := make(map[string][]string)
	for i, l := range lines {
		trimmed := strings.TrimSpace(l)
		if len(trimmed) == 0 {
			continue
		}
		switch trimmed {
		case GIVEN, WHEN, THEN, ERROR, EVENTS:
			current = trimmed
			continue
		default:
			if len(current) == 0 {
				return nil, fmt.Errorf("Expected suffix on line %d, got: %s", i, l)
			}
			array, _ := result[current]
			array = append(array, trimmed)
			result[current] = array
		}

	}
	return result, nil
}

const (
	GIVEN  = "GIVEN:"
	WHEN   = "WHEN:"
	THEN   = "THEN:"
	ERROR  = "ERROR:"
	EVENTS = "EVENTS:"
)

func SpecFromParseableString(s string) (*api.Spec, error) {

	scanner := bufio.NewScanner(strings.NewReader(s))

	var lines []string
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if len(s) > 0 {
			lines = append(lines, scanner.Text())
		}
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	spec := &api.Spec{}

	spec.Name = lines[0]

	if !strings.HasPrefix(lines[1], "---") {
		return nil, fmt.Errorf("expected ----- bot got: %s", lines[1])
	}

	groups, err := groupLines(lines[2:])
	if err != nil {
		return nil, err
	}

	if lines, ok := groups[GIVEN]; ok {
		spec.Given, err = stringsToMsgs(lines)
		if err != nil {
			return nil, err
		}
	}
	if lines, found := groups[EVENTS]; found {
		spec.ThenEvents, err = stringsToMsgs(lines)
		if err != nil {
			return nil, err
		}
	}

	if lines, found := groups[WHEN]; found {
		if len(lines) > 1 {
			return nil, fmt.Errorf("must be only one message in %s. Got %v", WHEN, lines)
		}
		spec.When, err = stringToMsg(lines[0])
		if err != nil {
			return nil, err
		}
	}
	if lines, found := groups[THEN]; found {
		if len(lines) > 1 {
			return nil, fmt.Errorf("must be only one message in %s. Got %d", THEN, len(lines))
		}
		spec.ThenResponse, err = stringToMsg(lines[0])
		if err != nil {
			return nil, err
		}
	}
	if lines, found := groups[ERROR]; found {
		if len(lines) > 1 {
			return nil, fmt.Errorf("must be only one message in %s. Got %v", ERROR, lines)
		}

		if spec.ThenError, err = parseStatus(lines[0]); err != nil {
			return nil, fmt.Errorf("can't parse error: %w", err)
		}
	}

	return spec, nil
}

func statusToString(st *status.Status) string {
	return fmt.Sprintf("%s: %s", st.Code().String(), st.Message())
}

func parseStatus(s string) (*status.Status, error) {
	segments := strings.SplitN(s, ":", 2)

	first := strings.TrimSpace(segments[0])

	if code, ok := strToCode[first]; ok {
		// got a match
		errStr := ""
		if len(segments) > 1 {
			errStr = strings.TrimSpace(segments[1])

		}
		return status.New(code, errStr), nil
	} else {
		return nil, fmt.Errorf("error code must be a valid gRPC status. Got %s", segments[0])
	}
}
