BINDIR=bin
BINARYNAME=server
build:
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(BINARYNAME) ./cmd/$(BINARYNAME)/