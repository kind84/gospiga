GO_TEST_FLAGS ?= -race

# set pkgs to all packages
PKGS = ./...

# verbose mode
ifdef VERBOSE
	GO_TEST_FLAGS += -v
	GO_BUILD_FLAGS += -v
endif

default: build

go-generate:
	go generate $(PKGS)

build: go-generate
	go build $(GO_BUILD_FLAGS) -ldflags "all=$(GO_LDFLAGS)" $(PKGS)

test: go-generate
	go test $(GO_TEST_FLAGS) -ldflags "all=$(GO_LDFLAGS)" $(PKGS)

docker: docker-build docker-run

build-dependencies:
	docker build -t dependencies -f ./dependencies.Dockerfile .

docker-build: build-dependencies
	docker-compose build

docker-run:
	docker-compose up
