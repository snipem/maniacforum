# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOINSTALL=$(GOCMD) install -v
GOGET=$(GOCMD) get
BINARY_NAME=maniacforum
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build install
build:
		$(GOBUILD) -o $(BINARY_NAME) -v
test:
		$(GOTEST) -v ./...
clean:
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		$(GOBUILD) -o $(BINARY_NAME)
		./$(BINARY_NAME)
install:
		$(GOINSTALL) .
deps:
		$(GOGET) github.com/skratchdot/open-golang/open
		$(GOGET) github.com/gizak/termui
