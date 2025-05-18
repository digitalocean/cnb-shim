.EXPORT_ALL_VARIABLES:

.PHONY: test \
        build \
        clean \
        package \
        release

SHELL=/usr/bin/env bash -o pipefail

GO111MODULE := on

VERSION := "v0.3+do.2"

build:
	@GOOS=linux GOARCH=amd64 go build -o "bin/release" ./cmd/release/...

test:
	go test ./... -v

clean:
	-rm -f cnb-shim-$(VERSION).tgz
	-rm -f bin/release

package: clean build
	@tar cvzf cnb-shim-$(VERSION).tgz bin/ sbin/ README.md LICENSE

release:
	@git tag $(VERSION)
	@git push --tags origin master
