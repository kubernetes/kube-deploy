# kernel-style V=1 build verbosity
ifeq ("$(origin V)", "command line")
       BUILD_VERBOSE = $(V)
endif

ifeq ($(BUILD_VERBOSE),1)
       Q =
else
       Q = @
endif

VERSION=$(shell git describe --dirty)
REPO=github.com/coreos/container-linux-config-transpiler
LD_FLAGS="-w -X $(REPO)/internal/version.Raw=$(VERSION)"

export GOPATH=$(shell pwd)/gopath
export CGO_ENABLED:=0

.PHONY: all
all: bin/ct

gopath:
	$(Q)mkdir -p gopath/src/github.com/coreos
	$(Q)ln -s ../../../.. gopath/src/$(REPO)

.PHONY: test
test:
	$(Q)./test

.PHONY: vendor
vendor:
	$(Q)glide update --strip-vendor
	$(Q)glide-vc --use-lock-file --no-tests --only-code

.PHONY: clean
clean:
	$(Q)rm -rf bin

.PHONY: release
release: \
	bin/ct-$(VERSION)-x86_64-unknown-linux-gnu \
	bin/ct-$(VERSION)-x86_64-apple-darwin \
	bin/ct-$(VERSION)-x86_64-pc-windows-gnu.exe

bin/ct-%-x86_64-unknown-linux-gnu: GOARGS = GOOS=linux GOARCH=amd64
bin/ct-%-x86_64-apple-darwin: GOARGS = GOOS=darwin GOARCH=amd64
bin/ct-%-x86_64-pc-windows-gnu.exe: GOARGS = GOOS=windows GOARCH=amd64

bin/%: | gopath
	$(Q)$(GOARGS) go build -o $@ -ldflags $(LD_FLAGS) $(REPO)/internal
