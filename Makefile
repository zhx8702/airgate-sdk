# AirGate SDK Makefile

GO := GOTOOLCHAIN=local go

.PHONY: help lint fmt test proto clean

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ===================== 质量检查 =====================

lint: ## 代码检查
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "未安装 golangci-lint，回退到 go vet"; \
		$(GO) vet ./...; \
	fi
	@echo "代码检查通过"

fmt: ## 格式化代码
	@if command -v goimports > /dev/null 2>&1; then \
		goimports -w -local github.com/DouDOU-start .; \
	else \
		$(GO) fmt ./...; \
	fi
	@echo "代码格式化完成"

test: ## 运行测试
	@$(GO) test ./...
	@echo "测试完成"

# ===================== 代码生成 =====================

proto: ## 重新生成 protobuf 代码
	@cd proto && protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		plugin.proto
	@echo "Proto 代码生成完成"

# ===================== 清理 =====================

clean: ## 清理构建产物
	@$(GO) clean ./...
