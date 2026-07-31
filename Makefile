.PHONY: run test tidy

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	export GOPROXY=https://goproxy.cn,direct && go mod tidy
