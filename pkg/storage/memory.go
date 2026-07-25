package storage

import (
	"sync"

	"github.com/kubeinsights/kubeinsights/pkg/trace"
)

type MemoryStore struct {
	mu     sync.RWMutex
	traces []trace.Result
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (s *MemoryStore) Save(result trace.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.traces = append([]trace.Result{result}, s.traces...)
	if len(s.traces) > 1000 {
		s.traces = s.traces[:1000]
	}
}

func (s *MemoryStore) List(limit int) []trace.Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.traces) {
		limit = len(s.traces)
	}
	out := make([]trace.Result, limit)
	copy(out, s.traces[:limit])
	return out
}

func (s *MemoryStore) Get(traceID uint64) (trace.Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, result := range s.traces {
		if result.TraceID == traceID {
			return result, true
		}
	}
	return trace.Result{}, false
}
