APP := butterfly-monitor
CMD_DIR := ./cmd
BIN_DIR := bin

.PHONY: build run vet fmt tidy vendor test docker clean

# 构建：输出到 bin/，须在项目根目录执行（configs/ 与 logs/ 相对 CWD）
build:
	go build -o $(BIN_DIR)/$(APP) $(CMD_DIR)

# 运行：须在项目根目录执行
run:
	go run $(CMD_DIR)

vet:
	go vet ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

vendor:
	go mod vendor

test:
	go test ./...

docker:
	docker build -t $(APP) .

clean:
	rm -rf $(BIN_DIR)
