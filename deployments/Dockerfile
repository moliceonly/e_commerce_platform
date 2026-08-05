# 多阶段构建（基础镜像走 DaoCloud 代理，避免直连 Docker Hub 超时）
FROM docker.m.daocloud.io/library/golang:1.24-alpine AS build
WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/server

FROM docker.m.daocloud.io/library/alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
COPY --from=build /out/server /app/server
# LoadYAML 读相对路径 configs/config.{APP_ENV}.yaml（compose 使用 APP_ENV=docker）
COPY configs/config.docker.yaml /app/configs/
RUN mkdir -p /app/data/uploads
EXPOSE 8080
HEALTHCHECK CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["/app/server"]
