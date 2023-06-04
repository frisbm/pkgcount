.PHONY: fmt lint build

fmt:
	gofmt -s -w .

lint:
	golangci-lint run

build:
	go build
	mv pkgcount ~/golang/bin/pkgcount

validate-tag-arg:
ifeq ("", "$(v)")
	@echo "version arg (v) must be used with the 'tag' target"
	@exit 1;
endif
ifneq ("v", "$(shell echo $(v) | head -c 1)")
	@echo "version arg (v) must begin with v"
	@exit 1;
endif

# ex: make tag v=v0.1.0
tag: validate-tag-arg
	@echo "creating tag $(v)"
	git tag $(v)
	git push origin $(v)
