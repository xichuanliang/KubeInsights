.PHONY: test run-once run-api

test:
	go test ./...

run-once:
	go run ./cmd/agent --once --url https://api.example.com/order

run-api:
	go run ./cmd/agent --listen :8080

