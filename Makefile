APP := codex-heartbeat
CMD := ./cmd/codex-heartbeat
INSTALL_DIR ?= $(HOME)/go/bin

.PHONY: build test install

build:
	go build -o ./$(APP) $(CMD)

test:
	go test ./...

install:
	mkdir -p $(INSTALL_DIR)
	GOBIN="$(INSTALL_DIR)" go install $(CMD)
