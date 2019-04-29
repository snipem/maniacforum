# maniacforum parameters
BINARY_NAME=maniacforum

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
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
		MANIACFORUM_READLOG_FILE=/tmp/maniac_read go run maniacforum.go
	
ui:
		tmux send-keys -t right "C-c"; sleep 0.1; tmux send-keys -t right "cd ${GOPATH}/src/github.com/snipem/maniacforum && make run" "C-m"; tmux select-pane -t right

run_binary:
		$(GOBUILD) -o $(BINARY_NAME)
		./$(BINARY_NAME)
install:
		$(GOINSTALL) .
deps:
		$(GOGET) github.com/skratchdot/open-golang/open
		$(GOGET) github.com/gizak/termui
		$(GOGET) github.com/PuerkitoBio/goquery
		$(GOGET) github.com/stretchr/testify/assert

release:
		git tag $(TAG_VERSION)
		goreleaser release --rm-dist
