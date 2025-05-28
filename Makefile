GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

REQUIREMENTS = "docker go"

# Default test file
TEST ?= examples/advanced.js

# Default target
.DEFAULT_GOAL := help

.PHONY: help build run-docker docker-up docker-down test test-k6 run-all clean requirements unit-test coverage

test: coverage ## Run Go unit tests and coverage report

test-k6: build docker-up ## Run the k6 test with the built binary (uses docker)[[BR]]Syntax:  `make test TEST=path/to/test.js ARGS="-v"`[[BR]]Default: TEST=examples/advanced.js
	@./k6 run $(TEST) $(ARGS)

run-all: docker-up test test-k6 docker-down ## Build the k6 binary, start Docker containers, run all tests, and then stop Docker containers[[BR]]Syntax: `make run-all TEST=path/to/test.js ARGS="-v"`

build: requirements ## Build the k6 binary
	@if [ ! -f k6 ]; then \
        if ! command -v xk6 &> /dev/null; then \
            make build-docker; \
        else \
            make build-local; \
        fi; \
    fi

build-local: # Build the k6 binary using the local xk6
	@xk6 build --with github.com/Hyperspace-Metaverse/xk6-socketio=.

build-docker: # Build the k6 binary using Docker image
	@docker run --rm -it -e GOOS=$(GOOS) -e GOARCH=$(GOARCH) -u "$(id -u):$(id -g)" -v "${PWD}:/xk6" grafana/xk6 build --with github.com/Hyperspace-Metaverse/xk6-socketio=.

docker-up: requirements ## Up the container to run k6 tests
	@docker compose up -d

docker-down: ## Stop the Docker container
	@docker compose down

clean: docker-down ## Clean all build/test artifacts and stops docker
	@[ -f k6 ] && rm -f k6 || true
	@[ -d test/node_modules ] && rm -rf test/node_modules || true

requirements:
	@for requirement in "$(REQUIREMENTS)"; do command -v $$requirement > /dev/null || (echo "\nERROR: \033[36m$$requirement\033[0m is not installed\n"; exit 1); done;

help: ## Show this help message and syntax for each target
	@awk '\
	  BEGIN {FS = ":.*##"; maxlen=0; linecount=0; print "\nUsage: make \033[36mtask\033[0m [VARIABLE=value ...]\n"} \
	  { if ($$0 != "") { linecount++; lines[linecount]=$$0 } } \
	  /^[a-zA-Z0-9_-]+:.*##/ { if (length($$1) > maxlen) maxlen = length($$1) } \
	  END { \
	    padlen = maxlen + 9; \
	    for (i=1; i<=linecount; i++) { \
	      line = lines[i]; \
	      if (match(line, /^[a-zA-Z0-9_-]+:.*##/)) { \
	        split(line, arr, ":.*##"); \
	        target = arr[1]; \
	        helpMsg = substr(line, RLENGTH+1); \
	        n = split(helpMsg, msgLines, "\\[\\[BR\\]\]"); \
	        printf "    \033[36m%-*s\033[0m    %s\n", maxlen, target, msgLines[1]; \
	        pad = sprintf("%*s", padlen, ""); \
	        for (j=2; j<=n; j++) { \
	          if (msgLines[j] ~ /^ *(Syntax:|Default:)/) { \
	            printf "%s\033[90m%s\033[0m\n", pad, msgLines[j]; \
	          } else { \
	            printf "%s%s\n", pad, msgLines[j]; \
	          } \
	        } \
	        print ""; \
	      } \
	    } \
	  }' $(MAKEFILE_LIST)

unit-test: ## Run Go unit tests with coverage
	go test -v -coverprofile=coverage.out ./...

coverage: unit-test ## Show Go test coverage report
	go tool cover -func=coverage.out