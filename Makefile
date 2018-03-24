TEST?=./...
NAME = $(shell awk -F\" '/^const appName/ { print $$2 }' main.go)
VERSION = $(shell awk -F\" '/^const appVer/ { print $$2 }' main.go)
DEPS = $(shell go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)

all: deps build

deps:
	go get -d -v ./...
	echo $(DEPS) | xargs -n1 go get -d

build: deps
	@mkdir -p bin/
	go build -o bin/$(NAME)

test:
	go test $(TEST) $(TESTARGS) -timeout=30s -parallel=4
	go test $(TEST) -race
	go vet $(TEST)

xcompile: deps
	@rm -rf build/
	@mkdir -p build
	gox \
		-output="build/{{.Dir}}_$(VERSION)_{{.OS}}_{{.Arch}}/$(NAME)"

package: xcompile
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done
	@mkdir -p build/zipfiles
	for f in $(FILES); do \
		(cd $(shell pwd)/build && zip zipfiles/$$f.zip $$f/*); \
		echo $$f; \
	done

.PHONY: all deps build test xcompile package
