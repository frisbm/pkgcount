
.PHONY: help
# help:
#    Print this help message
help:
	@grep -o '^\#.*' Makefile | cut -d" " -f2-

.PHONY: fmt
# fmt:
#    Format go code
fmt:
	goimports -local github.com/frisbm -w ./

.PHONY: lint
# lint:
#    Lint the code
lint:
	golangci-lint run

.PHONY: build
# build:
#    Build and install the binary
build:
	go build -o ./tmp .
	rm -rf "$$GOPATH/bin/pkgcount"
	go install .

.PHONY: tag
# tag:
#    Create a tag for the commits since the last tag and push it to the remote
tag:
	@echo "creating tag"
	bash ./scripts/tag.sh
