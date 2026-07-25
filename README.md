# KubeInsights

KubeInsights is an eBPF-oriented HTTP/HTTPS network path and performance diagnosis system scaffolded from `designer.md`.

The current implementation provides:

- A Go agent entry point in `cmd/agent`.
- A unified event model with trace ID, socket cookie, PID, interface, IPs, event type, and metadata.
- A mock RingBuffer-like collector flow that emits TCP, TLS, HTTP, network path, application, dependency, response, and close events.
- Trace aggregation and root cause analysis.
- A small HTTP API for health, traces, and local network topology.
- CO-RE/libbpf C source scaffolding under `bpf/`.
- Kubernetes deployment manifests under `deploy/kubernetes/`.

## Run locally

```bash
go run ./cmd/agent --once --url https://api.example.com/order
```

Start the API:

```bash
go run ./cmd/agent --listen :8080
```

Endpoints:

- `GET /healthz`
- `GET /api/traces`
- `GET /api/traces/{traceId}`
- `GET /api/topology`

## Test

```bash
go test ./...
```

## eBPF status

The `bpf/*.bpf.c` files define the event ABI and example kprobe handlers that submit events through a BPF ring buffer. The Windows development environment can compile and test the Go pipeline, while real eBPF loading must be completed on a Linux host with clang, libbpf, bpftool, and suitable kernel headers.

