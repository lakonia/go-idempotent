.PHONY: all
all: clean build test

ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")
ALL_PACKAGES_FOR_LINT=$(shell go list ./... | grep -v "vendor" | sed -e "s/github.com\/lakonia\/go-idempotent//g" | sed -e "s/^\///g")

setup:
	go get -u golang.org/x/lint/golint

clean:
	go clean

compile:
	mkdir -p out/
	go build

build: compile fmt vet lint

vendor: go mod vendor

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

lint:
	@for p in $(ALL_PACKAGES_FOR_LINT); do \
		echo "==> Linting $$p"; \
		golint $$p | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" | grep -vwE "comment on exported (var|function|method|type|const) \S+ should be of the form" || true; } \
	done

test:
	@echo "> running test"
	ENVIRONMENT=test go test $(ALL_PACKAGES) -race --v
