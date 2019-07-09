BINDIR := ./bin
CMDDIR := ./cmd
CMDS := $(notdir $(wildcard $(CMDDIR)/*))
BINS := $(addprefix $(BINDIR)/,$(CMDS))
.PHONY: all test clean install

all: $(BINS)

$(BINDIR)/%: */**/*.go bootstrap
	@mkdir -p $(BINDIR)
	go build -o $@ -ldflags="-X main.buildVersion=$(shell $(BINDIR)/git-semver-describe --tags --trim=v)" $(CMDDIR)/$*

test:
	go test ./... -race

clean:
	rm -rf $(BINDIR)
	rm -rf dist

# we require a copy of ourselves to build ourselves with version info (self hosting lol),
# so if no copy already exists, bootstrap one without version information in order to make the first build
bootstrap:
	@mkdir -p $(BINDIR)
	@if [ ! -f "$(BINDIR)/git-semver-describe" ]; then \
		echo "bootstrapping with unversioned copy of git-semver-describe"; \
		go build -o $(BINDIR)/git-semver-describe $(CMDDIR)/git-semver-describe; \
	fi

install:
	go install -ldflags="-X main.buildVersion=$(shell $(BINDIR)/git-semver-describe --tags --trim=v)" $(CMDDIR)/git-semver-describe

snapshot:
	VERSION=$(shell $(BINDIR)/git-semver-describe --tags --trim=v) goreleaser --snapshot --rm-dist
