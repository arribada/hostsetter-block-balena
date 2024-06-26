GO111MODULE 	?= on
export GO111MODULE

GOPATH			?= $(shell go env GOPATH)

# otoh, since we're building from the same place, when we're switching arch,
# since we're going to remove the intermediate build objects, we don't want to
# overwite the binaries, so give them different names.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    HOSTSETTER ?= devhostsetter
else
    HOSTSETTER ?= hostsetter
endif

.PHONY: all
all: build

.PHONY: clean
clean: ## clean.
clean:
	@rm ./bin/${HOSTSETTER}

.PHONY: build
build: ## Build.
build: ./bin/${HOSTSETTER}
	@echo ">> Building"

.PHONY: ./bin/${HOSTSETTER}
./bin/${HOSTSETTER}:
	@mkdir -p bin
	@go build -o ./bin/${HOSTSETTER} ./cmd/hostsetter/main.go

.PHONY: prelint
prelint: ## Prepare for linting.
prelint: build
	@echo ">> Prepare deps for linting"
