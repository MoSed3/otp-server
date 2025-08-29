NAME = server-$(GOOS)-$(GOARCH)
CLI_NAME = admin-cli-$(GOOS)-$(GOARCH)

LDFLAGS = -s -w -buildid=
PARAMS = -trimpath -ldflags "$(LDFLAGS)" -v
MAIN = ./cmd/server
CLI_MAIN = ./cmd/admin/main.go
PREFIX ?= $(shell go env GOPATH)

ifeq ($(GOOS),windows)
OUTPUT = $(NAME).exe
CLI_OUTPUT = $(CLI_NAME).exe
ADDITION = go build -o w$(NAME).exe -trimpath -ldflags "-H windowsgui $(LDFLAGS)" -v $(MAIN)
else
OUTPUT = $(NAME)
CLI_OUTPUT = $(CLI_NAME)
endif

ifeq ($(shell echo "$(GOARCH)" | grep -Eq "(mips|mipsle)" && echo true),true)
ADDITION = GOMIPS=softfloat go build -o $(NAME)_softfloat -trimpath -ldflags "$(LDFLAGS)" -v $(MAIN)
endif

.PHONY: clean build build-cli
build:
	CGO_ENABLED=0 go build -o $(OUTPUT) $(PARAMS) $(MAIN)
	$(ADDITION)

build-cli:
	CGO_ENABLED=0 go build -o $(CLI_OUTPUT) $(PARAMS) $(CLI_MAIN)

clean:
	go clean -v -i $(PWD)
	rm -f $(NAME)-* w$(NAME)-*.exe $(CLI_NAME)-*

deps:
	go mod download
	go mod tidy

run:
	go run ./cmd/server

ARGS = $(filter-out $@,$(MAKECMDGOALS))
.PHONY: run-cli
run-cli:
	go run ./cmd/admin $(ARGS)

swag:
	swag init -g cmd/server/main.go

swag-install:
	go install github.com/swaggo/swag/cmd/swag@latest
