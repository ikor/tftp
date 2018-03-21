BINDIR=./cmd/server
BINARYNAME=ramtftp

.PHONY: all
all: build

.PHONY: build
build:
	go build -o $(BINDIR)/$(BINARYNAME) $(BINDIR)

.PHONY: test
test:
	go test -cover -v ./tftp/
