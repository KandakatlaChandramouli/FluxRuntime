package exporter

import (
	"encoding/csv"
	"os"
)

func WriteCSV(path string, rows [][]string) error {
	f, err := os.Create(path)

	if err != nil {
		return err
	}

	defer f.Close()

	w := csv.NewWriter(f)

	defer w.Flush()

	return w.WriteAll(rows)
}
