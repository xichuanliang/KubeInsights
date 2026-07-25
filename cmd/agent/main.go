package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubeinsights/kubeinsights/pkg/analyzer"
	"github.com/kubeinsights/kubeinsights/pkg/api"
	"github.com/kubeinsights/kubeinsights/pkg/collector"
	"github.com/kubeinsights/kubeinsights/pkg/storage"
	"github.com/kubeinsights/kubeinsights/pkg/topology"
	"github.com/kubeinsights/kubeinsights/pkg/trace"
)

func main() {
	var (
		mode   = flag.String("mode", "mock", "collector mode: mock or ebpf")
		url    = flag.String("url", "https://api.example.com/order", "URL used by the mock collector")
		listen = flag.String("listen", ":8080", "HTTP API listen address")
		once   = flag.Bool("once", false, "collect one trace, print JSON, and exit")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	src, err := newCollector(*mode, *url)
	if err != nil {
		logger.Error("create collector", "error", err)
		os.Exit(1)
	}

	store := storage.NewMemoryStore()
	engine := trace.NewEngine(analyzer.DefaultRules())

	if *once {
		result, err := collectOnce(ctx, src, engine)
		if err != nil {
			logger.Error("collect trace", "error", err)
			os.Exit(1)
		}
		writeJSON(result)
		return
	}

	go runCollector(ctx, logger, src, engine, store)

	server := &http.Server{
		Addr:              *listen,
		Handler:           api.NewServer(store, topology.NewDiscoverer()).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("api listening", "addr", *listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("api server stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func newCollector(mode, url string) (collector.Source, error) {
	switch mode {
	case "mock":
		return collector.NewMockSource(url), nil
	case "ebpf":
		return collector.NewEBPFSource(), nil
	default:
		return nil, fmt.Errorf("unknown collector mode %q", mode)
	}
}

func runCollector(ctx context.Context, logger *slog.Logger, src collector.Source, engine *trace.Engine, store *storage.MemoryStore) {
	events, err := src.Start(ctx)
	if err != nil {
		logger.Error("start collector", "error", err)
		return
	}
	for ev := range events {
		if result, ok := engine.Add(ev); ok {
			store.Save(result)
			logger.Info("trace completed", "trace_id", result.TraceID, "root_cause", result.RootCause, "duration_ms", result.DurationMS)
		}
	}
}

func collectOnce(ctx context.Context, src collector.Source, engine *trace.Engine) (trace.Result, error) {
	events, err := src.Start(ctx)
	if err != nil {
		return trace.Result{}, err
	}
	for ev := range events {
		if result, ok := engine.Add(ev); ok {
			return result, nil
		}
	}
	return trace.Result{}, context.Canceled
}

func writeJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
}
