package specs

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/trustbit/bfkata/api"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"testing"
)

func TestMessageConversion(t *testing.T) {
	samples := []struct {
		msg proto.Message
		txt string
	}{
		{
			msg: &api.LocationAdded{Id: 1, Name: "loc", Parent: 2},
			txt: `LocationAdded id:1 name:"loc" parent:2`,
		}, {
			msg: &api.Reserved{Reservation: 2, Code: "ASD"},
			txt: `Reserved reservation:2 code: "ASD"`,
		},
	}

	for _, s := range samples {
		t.Run(s.txt, func(t *testing.T) {
			actual, err := stringToMsg(s.txt)
			if err != nil {
				t.Fatalf("parsing error: %s", err)
			}
			deltas := cmp.Diff(actual, s.msg, protocmp.Transform())
			if len(deltas) > 0 {
				t.Fatalf(deltas)
			}

		})
	}

}

func TestRoundtrip(t *testing.T) {
	for _, s := range api.Specs {
		t.Run(s.Name, func(t *testing.T) {

			f, err := SpecToParseableString(s)
			if err != nil {
				t.Fatalf("format: %s", err)
			}

			result, err := SpecFromParseableString(f)
			if err != nil {
				t.Fatalf("parse: %s", err)
			}

			if s.ToTestString() != result.ToTestString() {
				t.Fatalf(cmp.Diff(s.ToTestString(), result.ToTestString()))

			}

		})

	}
}

func TestFullRoundTrip(t *testing.T) {
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)

	err := WriteSpecs(api.Specs, foo)
	if err != nil {
		t.Fatal(err)
	}

	_ = foo.Flush()

	actual, err := ReadSpecs(&b)
	if err != nil {
		t.Log(b.String())
		t.Fatal(err)
	}

	if len(api.Specs) != len(actual) {
		t.Fatalf("Length mismatch: %d, %d", len(api.Specs), len(actual))
	}

	for i, s := range api.Specs {

		t.Run(fmt.Sprintf("#%d: %s", i, s.Name), func(t *testing.T) {

			delta := cmp.Diff(s.ToTestString(), actual[i].ToTestString())
			if len(delta) > 0 {
				t.Fatal(delta)
			}
		})
	}

}
