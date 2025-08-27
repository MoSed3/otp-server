NAME = otp-server-$(GOOS)-$(GOARCH)

LDFLAGS = -s -w -buildid=
PARAMS = -trimpath -ldflags "$(LDFLAGS)" -v
MAIN = ./main.go
PREFIX ?= $(shell go env GOPATH)

ifeq ($(GOOS),windows)
OUTPUT = $(NAME).exe
ADDITION = go build -o w$(NAME).exe -trimpath -ldflags "-H windowsgui $(LDFLAGS)" -v $(MAIN)
else
OUTPUT = $(NAME)
endif

ifeq ($(shell echo "$(GOARCH)" | grep -Eq "(mips|mipsle)" && echo true),true)
ADDITION = GOMIPS=softfloat go build -o $(NAME)_softfloat -trimpath -ldflags "$(LDFLAGS)" -v $(MAIN)
endif

.PHONY: clean build

build:
	CGO_ENABLED=0 go build -o $(OUTPUT) $(PARAMS) $(MAIN)
	$(ADDITION)

clean:
	go clean -v -i $(PWD)
	rm -f $(NAME)-* w$(NAME)-*.exe

deps:
	go mod download
	go mod tidy
