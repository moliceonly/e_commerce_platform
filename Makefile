.PHONY: run test tidy lint

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	export GOPROXY=https://goproxy.cn,direct && go mod tidy

lint:
	golangci-lint run ./...
