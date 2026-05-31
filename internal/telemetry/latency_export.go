package telemetry

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"
)

type LatencySample struct {
	TimestampNs int64
	LatencyNs   int64
	QueueDepth  int64
	WorkerID    int
}

type Exporter struct {
	mu sync.Mutex
	f  *os.File
	w  *csv.Writer
}

func NewExporter(path string) (*Exporter, error) {

	f, err := os.Create(path)

	if err != nil {
		return nil, err
	}

	w := csv.NewWriter(f)

	w.Write([]string{
		"timestamp_ns",
		"latency_ns",
		"queue_depth",
		"worker_id",
	})

	return &Exporter{
		f: f,
		w: w,
	}, nil
}

func (e *Exporter) Write(s LatencySample) {

	e.mu.Lock()
	defer e.mu.Unlock()

	e.w.Write([]string{
		strconv.FormatInt(s.TimestampNs, 10),
		strconv.FormatInt(s.LatencyNs, 10),
		strconv.FormatInt(s.QueueDepth, 10),
		strconv.Itoa(s.WorkerID),
	})

	e.w.Flush()
}

func (e *Exporter) Close() {
	e.w.Flush()
	e.f.Close()
}
