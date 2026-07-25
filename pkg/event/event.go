package event

import "time"

type Type string

const (
	TypeSocketConnect Type = "SOCKET_CONNECT"
	TypeTCPRTT        Type = "TCP_RTT"
	TypeTCPRetransmit Type = "TCP_RETRANSMIT"
	TypeNetworkHop    Type = "NETWORK_HOP"
	TypeHTTP          Type = "HTTP"
	TypeTLSHandshake  Type = "TLS_HANDSHAKE"
	TypeApplication   Type = "APPLICATION"
	TypeDependency    Type = "DEPENDENCY"
	TypeResponse      Type = "RESPONSE"
	TypeSocketClose   Type = "SOCKET_CLOSE"
)

type Event struct {
	Timestamp    time.Time         `json:"timestamp"`
	Type         Type              `json:"type"`
	TraceID      uint64            `json:"traceId"`
	SocketCookie uint64            `json:"socketCookie"`
	PID          uint32            `json:"pid"`
	TID          uint32            `json:"tid"`
	Interface    string            `json:"interface,omitempty"`
	SrcIP        string            `json:"srcIp,omitempty"`
	DstIP        string            `json:"dstIp,omitempty"`
	SrcPort      uint16            `json:"srcPort,omitempty"`
	DstPort      uint16            `json:"dstPort,omitempty"`
	Duration     time.Duration     `json:"duration"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

func New(t Type, traceID, socketCookie uint64, duration time.Duration) Event {
	return Event{
		Timestamp:    time.Now().UTC(),
		Type:         t,
		TraceID:      traceID,
		SocketCookie: socketCookie,
		PID:          uint32(1000 + traceID%5000),
		TID:          uint32(2000 + traceID%5000),
		Interface:    "eth0",
		SrcIP:        "10.244.1.10",
		DstIP:        "10.96.0.42",
		SrcPort:      43000 + uint16(traceID%1000),
		DstPort:      443,
		Duration:     duration,
		Metadata:     map[string]string{},
	}
}
