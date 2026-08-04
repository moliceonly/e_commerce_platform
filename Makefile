.PHONY: run test tidy lint

run:
	go run ./cmd/server

test:
	go test ./...

tidy:
	export GOPROXY=https://goproxy.cn,direct && go mod tidy

# 阶段 H4：需本机安装 golangci-lint
lint:
	golangci-lint run ./...
