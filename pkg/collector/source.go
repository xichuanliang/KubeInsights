package collector

import (
	"context"
	"errors"
	"time"

	"github.com/kubeinsights/kubeinsights/pkg/event"
)

type Source interface {
	Start(context.Context) (<-chan event.Event, error)
}

type MockSource struct {
	URL string
}

func NewMockSource(url string) *MockSource {
	return &MockSource{URL: url}
}

func (s *MockSource) Start(ctx context.Context) (<-chan event.Event, error) {
	out := make(chan event.Event)
	go func() {
		defer close(out)
		traceID := uint64(time.Now().UnixNano())
		socketCookie := traceID ^ 0xC0FFEE
		events := []event.Event{
			withMeta(event.New(event.TypeSocketConnect, traceID, socketCookie, 5*time.Millisecond), "phase", "tcp_connect"),
			withMeta(event.New(event.TypeTCPRTT, traceID, socketCookie, 3*time.Millisecond), "rtt_ms", "3"),
			withMeta(event.New(event.TypeTLSHandshake, traceID, socketCookie, 20*time.Millisecond), "version", "TLS1.3"),
			withMeta(event.New(event.TypeHTTP, traceID, socketCookie, 1*time.Millisecond), "method", "GET", "url", s.URL, "host", "api.example.com"),
			withMeta(event.New(event.TypeNetworkHop, traceID, socketCookie, 3*time.Millisecond), "device", "eth0", "ip", "10.244.1.10", "namespace", "pod/order"),
			withMeta(event.New(event.TypeNetworkHop, traceID, socketCookie, 5*time.Millisecond), "device", "cni0", "ip", "10.244.0.1", "namespace", "root"),
			withMeta(event.New(event.TypeApplication, traceID, socketCookie, 50*time.Millisecond), "service", "order-service"),
			withMeta(event.New(event.TypeDependency, traceID, socketCookie, 400*time.Millisecond), "kind", "mysql", "statement", "SELECT order"),
			withMeta(event.New(event.TypeResponse, traceID, socketCookie, 2*time.Millisecond), "status", "200"),
			withMeta(event.New(event.TypeSocketClose, traceID, socketCookie, 0), "phase", "tcp_close"),
		}
		for _, ev := range events {
			select {
			case <-ctx.Done():
				return
			case out <- ev:
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
	return out, nil
}

type EBPFSource struct{}

func NewEBPFSource() *EBPFSource {
	return &EBPFSource{}
}

func (s *EBPFSource) Start(context.Context) (<-chan event.Event, error) {
	return nil, errors.New("ebpf collector is scaffolded; build bpf/*.bpf.c and wire libbpf ring buffer on a Linux host")
}

func withMeta(ev event.Event, kv ...string) event.Event {
	for i := 0; i+1 < len(kv); i += 2 {
		ev.Metadata[kv[i]] = kv[i+1]
	}
	return ev
}
