package export

import (
	"encoding/json"
	"os"
)

func Write(path string, v any) error {
	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	enc := json.NewEncoder(f)

	return enc.Encode(v)
}
