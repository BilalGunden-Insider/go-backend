package worker

import "sync/atomic"

type Stats struct {
	Processed atomic.Int64
	Failed    atomic.Int64
	Pending   atomic.Int64
}

func (s *Stats) Snapshot() StatsSnapshot {
	return StatsSnapshot{
		Processed: s.Processed.Load(),
		Failed:    s.Failed.Load(),
		Pending:   s.Pending.Load(),
	}
}

type StatsSnapshot struct {
	Processed int64 `json:"processed"`
	Failed    int64 `json:"failed"`
	Pending   int64 `json:"pending"`
}
