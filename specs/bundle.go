package specs

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/trustbit/bfkata/api"
	"os"
)

//go:embed bundle.txt
var BundledSpecs string

const BUNDLE = "<bundle>"

func Load(file string) ([]*api.Spec, error) {
	var reader *bytes.Reader
	if file == BUNDLE {
		reader = bytes.NewReader([]byte(BundledSpecs))

	} else {
		in, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("can't read file: %w", err)
		}
		reader = bytes.NewReader(in)
	}
	actual, err := ReadSpecs(reader)
	if err != nil {
		return nil, fmt.Errorf("can't parse specs:", err)
	}
	return actual, nil

}
