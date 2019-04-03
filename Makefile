# maniacforum parameters
BINARY_NAME=maniacforum
TAG_VERSION=0.0.4

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOINSTALL=$(GOCMD) install -v
GOGET=$(GOCMD) get
BINARY_NAME=maniacforum

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

release:
		git tag $(TAG_VERSION)
		goreleaser release --rm-dist
