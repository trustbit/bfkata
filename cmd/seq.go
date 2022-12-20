package cmd

import (
	"fmt"
	"github.com/abdullin/go-seq"
	"github.com/trustbit/bfkata/api"
	"google.golang.org/protobuf/proto"
	"log"
)

func intToSeq(i int64) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012x", i)
}
func seqToInt(s string) int64 {
	i, ok := seq.ParseActualUid(s)
	if !ok {
		log.Panicln("Failed to parse uid: ", s)
	}
	return i
}

func nextSeq(msgs []proto.Message) string {

	var id int64

	inc := func(s string) {
		parsed := seqToInt(s)
		if parsed != id+1 {
			log.Panicln("Seq is not incrementing")
		}
		if parsed > id {
			id = parsed
		}
		id = parsed

	}
	for _, e := range msgs {
		switch t := e.(type) {
		case *api.LocationAdded:
			inc(t.Uid)
		case *api.Reserved:
			inc(t.Reservation)
		case *api.ProductAdded:
			inc(t.Uid)

		}
	}
	nextId := id + 1
	return intToSeq(nextId)
}
