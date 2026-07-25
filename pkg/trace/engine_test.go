package trace

import (
	"testing"
	"time"

	"github.com/kubeinsights/kubeinsights/pkg/analyzer"
	"github.com/kubeinsights/kubeinsights/pkg/event"
)

func TestBuildResultDetectsMySQLLatency(t *testing.T) {
	traceID := uint64(42)
	socketCookie := uint64(7)
	events := []event.Event{
		event.New(event.TypeSocketConnect, traceID, socketCookie, 5*time.Millisecond),
		withMeta(event.New(event.TypeHTTP, traceID, socketCookie, time.Millisecond), "url", "https://api.example.com/order"),
		withMeta(event.New(event.TypeDependency, traceID, socketCookie, 150*time.Millisecond), "kind", "mysql"),
		event.New(event.TypeSocketClose, traceID, socketCookie, 0),
	}

	result := BuildResult(events, analyzer.DefaultRules())
	if result.RootCause != "MYSQL_LATENCY" {
		t.Fatalf("root cause = %s, want MYSQL_LATENCY", result.RootCause)
	}
	if result.URL != "https://api.example.com/order" {
		t.Fatalf("url = %s", result.URL)
	}
}

func withMeta(ev event.Event, kv ...string) event.Event {
	for i := 0; i+1 < len(kv); i += 2 {
		ev.Metadata[kv[i]] = kv[i+1]
	}
	return ev
}
