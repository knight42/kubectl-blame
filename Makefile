NAME ?= kubectl-blame
VERSION ?= $(shell git describe --tags || echo "unknown")
GO_LDFLAGS='-w -s'
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags $(GO_LDFLAGS)

PLATFORM_LIST = \
	darwin-amd64 \
	linux-amd64

all: linux-amd64 darwin-amd64 # Most used

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(NAME)-$(VERSION)-$@/$(NAME)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(NAME)-$(VERSION)-$@/$(NAME)

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

install:
	CGO_ENABLED=0 go install -trimpath -ldflags $(GO_LDFLAGS)

gz_releases=$(addsuffix .tar.gz, $(PLATFORM_LIST))

$(gz_releases): %.tar.gz : %
	tar czf $(NAME)-$(VERSION)-$@ $(NAME)-$(VERSION)-$</

releases: $(gz_releases)
clean:
	rm -r $(NAME)-*
