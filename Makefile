# AirGate SDK Makefile

GO := GOTOOLCHAIN=local go

.PHONY: help ci pre-commit lint fmt test vet build proto clean setup-hooks

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ===================== 质量检查 =====================

ci: lint test vet build ## 本地运行与 CI 完全一致的检查

pre-commit: lint vet build ## pre-commit hook 调用（跳过耗时的 race 测试）

lint: ## 代码检查（需要安装 golangci-lint）
	@if ! command -v golangci-lint > /dev/null 2>&1; then \
		echo "错误: 未安装 golangci-lint，请执行: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	golangci-lint run ./...
	@echo "代码检查通过"

fmt: ## 格式化代码
	@if command -v goimports > /dev/null 2>&1; then \
		goimports -w -local github.com/DouDOU-start .; \
	else \
		$(GO) fmt ./...; \
	fi
	@echo "代码格式化完成"

test: ## 运行测试（race 检测 + 覆盖率）
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	@echo "测试完成"

vet: ## 静态分析
	$(GO) vet ./...

build: ## 编译检查
	$(GO) build ./...

# ===================== 代码生成 =====================

proto: ## 重新生成 protobuf 代码
	@cd proto && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		plugin.proto
	@echo "Proto 代码生成完成"

# ===================== Git Hooks =====================

setup-hooks: ## 安装 Git pre-commit hook
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "pre-commit hook 已安装"

# ===================== 清理 =====================

clean: ## 清理构建产物
	@$(GO) clean ./...
