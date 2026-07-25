package trace

import (
	"sort"
	"time"

	"github.com/kubeinsights/kubeinsights/pkg/analyzer"
	"github.com/kubeinsights/kubeinsights/pkg/event"
)

type Span struct {
	Name       event.Type        `json:"name"`
	DurationMS int64             `json:"durationMs"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type NetworkHop struct {
	Device    string `json:"device"`
	IP        string `json:"ip"`
	Namespace string `json:"namespace,omitempty"`
	LatencyMS int64  `json:"latencyMs"`
}

type Result struct {
	TraceID     uint64        `json:"traceId"`
	URL         string        `json:"url,omitempty"`
	DurationMS  int64         `json:"durationMs"`
	NetworkPath []NetworkHop  `json:"networkPath"`
	Spans       []Span        `json:"spans"`
	RootCause   string        `json:"rootCause"`
	Events      []event.Event `json:"events,omitempty"`
}

type Engine struct {
	rules  analyzer.Rules
	traces map[uint64][]event.Event
}

func NewEngine(rules analyzer.Rules) *Engine {
	return &Engine{
		rules:  rules,
		traces: map[uint64][]event.Event{},
	}
}

func (e *Engine) Add(ev event.Event) (Result, bool) {
	e.traces[ev.TraceID] = append(e.traces[ev.TraceID], ev)
	if ev.Type != event.TypeSocketClose && ev.Type != event.TypeResponse {
		return Result{}, false
	}
	events := e.traces[ev.TraceID]
	if ev.Type == event.TypeResponse && !hasType(events, event.TypeSocketClose) {
		return Result{}, false
	}
	delete(e.traces, ev.TraceID)
	return BuildResult(events, e.rules), true
}

func BuildResult(events []event.Event, rules analyzer.Rules) Result {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	result := Result{Events: events}
	for _, ev := range events {
		if result.TraceID == 0 {
			result.TraceID = ev.TraceID
		}
		if url := ev.Metadata["url"]; url != "" {
			result.URL = url
		}
		result.DurationMS += millis(ev.Duration)
		switch ev.Type {
		case event.TypeNetworkHop:
			result.NetworkPath = append(result.NetworkPath, NetworkHop{
				Device:    firstNonEmpty(ev.Metadata["device"], ev.Interface),
				IP:        firstNonEmpty(ev.Metadata["ip"], ev.SrcIP),
				Namespace: ev.Metadata["namespace"],
				LatencyMS: millis(ev.Duration),
			})
		case event.TypeSocketClose:
		default:
			result.Spans = append(result.Spans, Span{
				Name:       ev.Type,
				DurationMS: millis(ev.Duration),
				Metadata:   ev.Metadata,
			})
		}
	}
	result.RootCause = analyzer.Analyze(analysisSpans(result.Spans), rules)
	return result
}

func analysisSpans(spans []Span) []analyzer.Span {
	out := make([]analyzer.Span, 0, len(spans))
	for _, span := range spans {
		out = append(out, analyzer.Span{
			Name:       span.Name,
			DurationMS: span.DurationMS,
			Metadata:   span.Metadata,
		})
	}
	return out
}

func hasType(events []event.Event, t event.Type) bool {
	for _, ev := range events {
		if ev.Type == t {
			return true
		}
	}
	return false
}

func millis(d time.Duration) int64 {
	return d.Milliseconds()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
