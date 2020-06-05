VERSION = $(shell git describe --tags)
GITREV = $(shell git rev-parse --verify --short HEAD)
GITBRANCH = $(shell git rev-parse --abbrev-ref HEAD)
DATE = $(shell LANG=US date +"%a, %d %b %Y %X %z")
GOVERSION = 1.14.4

GO_LDFLAGS += -X 'gospiga.Version=$(VERSION)'
GO_LDFLAGS += -X 'gospiga.GitRev=$(GITREV)'
GO_LDFLAGS += -X 'gospiga.GitBranch=$(GITBRANCH)'
GO_LDFLAGS += -X 'gospiga.BuildDate=$(DATE)'

DOCKER_TAG = latest
ifdef GITHUB_REF
DOCKER_TAG = $(notdir $(GITHUB_REF))
endif

DGRAPH_TAG = v20.03.3
