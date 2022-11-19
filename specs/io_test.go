package specs

import (
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
