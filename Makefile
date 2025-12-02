.PHONY: build run test clean fmt lint deps help

# 项目信息
BINARY_NAME=otter
MAIN_PATH=./cmd/otter
BUILD_DIR=./bin

# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# 构建标志
LDFLAGS=-ldflags "-s -w"
BUILD_FLAGS=-trimpath

# 默认目标
.DEFAULT_GOAL := help

## build: 构建项目
build:
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "构建完成: $(BUILD_DIR)/$(BINARY_NAME)"

## run: 运行项目
run:
	@echo "运行 $(BINARY_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	$(BUILD_DIR)/$(BINARY_NAME)

## test: 运行测试
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@echo "运行测试并生成覆盖率报告..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## clean: 清理构建文件
clean:
	@echo "清理构建文件..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "清理完成"

## fmt: 格式化代码
fmt:
	@echo "格式化代码..."
	$(GOFMT) ./...
	@echo "格式化完成"

## lint: 代码检查（需要安装 golangci-lint）
lint:
	@echo "代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint 未安装，跳过检查"; \
		echo "安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## deps: 下载依赖
deps:
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "依赖下载完成"

## deps-update: 更新依赖
deps-update:
	@echo "更新依赖..."
	$(GOMOD) get -u ./...
	$(GOMOD) tidy
	@echo "依赖更新完成"

## install: 安装到 $GOPATH/bin
install:
	@echo "安装 $(BINARY_NAME)..."
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $$(go env GOPATH)/bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "安装完成: $$(go env GOPATH)/bin/$(BINARY_NAME)"

## help: 显示帮助信息
help:
	@echo "可用命令:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

