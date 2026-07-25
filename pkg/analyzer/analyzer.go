package analyzer

import "github.com/kubeinsights/kubeinsights/pkg/event"

type Span struct {
	Name       event.Type
	DurationMS int64
	Metadata   map[string]string
}

type Rules struct {
	SlowDependencyMS int64
	SlowNetworkMS    int64
	SlowTLSMS        int64
	SlowAppMS        int64
}

func DefaultRules() Rules {
	return Rules{
		SlowDependencyMS: 100,
		SlowNetworkMS:    50,
		SlowTLSMS:        100,
		SlowAppMS:        100,
	}
}

func Analyze(spans []Span, rules Rules) string {
	var max Span
	for _, span := range spans {
		if span.DurationMS > max.DurationMS {
			max = span
		}
	}

	switch {
	case max.Name == event.TypeDependency && max.DurationMS >= rules.SlowDependencyMS:
		if max.Metadata["kind"] == "mysql" {
			return "MYSQL_LATENCY"
		}
		return "DEPENDENCY_LATENCY"
	case max.Name == event.TypeNetworkHop && max.DurationMS >= rules.SlowNetworkMS:
		return "NETWORK_LATENCY"
	case max.Name == event.TypeTLSHandshake && max.DurationMS >= rules.SlowTLSMS:
		return "TLS_LATENCY"
	case max.Name == event.TypeApplication && max.DurationMS >= rules.SlowAppMS:
		return "APPLICATION_LATENCY"
	case max.Name == event.TypeTCPRetransmit:
		return "NETWORK_PACKET_LOSS"
	default:
		return "NO_OBVIOUS_BOTTLENECK"
	}
}
