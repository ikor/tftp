BINDIR=./cmd/server
BINARYNAME=ramtftp

.PHONY: all
all: test build

.PHONY: build
build:
	go build -o $(BINDIR)/$(BINARYNAME) $(BINDIR)

.PHONY: test
test:
	go test -race -cover -v github.com/ikor/tftp/tftp github.com/ikor/tftp/cmd/server
