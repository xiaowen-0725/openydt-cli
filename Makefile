BINARY  := openydt
MODULE  := github.com/xiaowen-0725/openydt-cli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(MODULE)/internal/cmdutil.Version=$(VERSION)

.PHONY: all catalog generate build vet test smoke e2e clean

all: build

## catalog: 从 open-api-front 的 Doc/*.vue 抽取接口目录 -> catalog/catalog.json
catalog:
	cd tools/extractor && npm install --no-audit --no-fund && node extract.mjs

## generate: 由 catalog.json 生成各域命令 -> cmd/gen/*.go
generate:
	go run ./internal/gen catalog/catalog.json cmd/gen
	gofmt -w cmd/gen

## build: 编译单二进制 -> bin/openydt
build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

vet:
	go vet ./...

test:
	go test ./...

## smoke: 对测试环境做查费冒烟(需先 config set 或设置 OPENYDT_KEY/SECRET)
smoke: build
	./bin/$(BINARY) auth test
	./bin/$(BINARY) api getParkFee --body '{"carCode":"粤EJW962"}' -o table

## e2e: 跑端到端场景验证, 产出 TEST_REPORT.md
e2e: build
	go run ./tests/e2e

clean:
	rm -rf bin
