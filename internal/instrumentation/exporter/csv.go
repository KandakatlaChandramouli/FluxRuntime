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

	err = w.WriteAll(rows)

	if err != nil {
		return err
	}

	w.Flush()

	return nil
}
