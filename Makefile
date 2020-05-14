include version.mk
GO_TEST_FLAGS ?= -race
SERVICES = server finder
REGISTRY = docker.pkg.github.com

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

docker: docker-server docker-finder

docker-server: build-dependencies
	docker build -t gospiga/server server

docker-finder: build-dependencies
	docker build -t gospiga/finder finder

build-dependencies:
	docker build -t dependencies -f ./dependencies.Dockerfile .

docker-build: build-dependencies
	docker-compose build

docker-run: docker-build
	docker-compose up

release: docker
	for service in $(SERVICES); do \
		docker tag gospiga/$$service $(REGISTRY)/$$service:$(DOCKER_TAG); \
		docker push $(REGISTRY)/$$service:$(DOCKER_TAG); \
	done
