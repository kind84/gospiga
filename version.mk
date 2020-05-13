VERSION = $(shell git describe --tags)
GITREV = $(shell git rev-parse --verify --short HEAD)
GITBRANCH = $(shell git rev-parse --abbrev-ref HEAD)
DATE = $(shell LANG=US date +"%a, %d %b %Y %X %z")

GO_LDFLAGS += -X 'ke.Version=$(VERSION)'
GO_LDFLAGS += -X 'ke.GitRev=$(GITREV)'
GO_LDFLAGS += -X 'ke.GitBranch=$(GITBRANCH)'
GO_LDFLAGS += -X 'ke.BuildDate=$(DATE)'

DOCKER_TAG = latest
ifdef GITHUB_REF
DOCKER_TAG = $(notdir $(GITHUB_REF))
endif
