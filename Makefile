SERVER_URL ?= https://host.docker.internal:9200

# Get the latest tag, or "dev" if none exists
TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.1")
# Get the short commit hash
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
# Check if current HEAD is exactly a tag
EXACT_TAG := $(shell git describe --tags --exact-match 2>/dev/null)

ifeq ($(EXACT_TAG),)
	# Not at a tag, append commit hash
	VERSION := $(TAG)-$(COMMIT)
else
	# At a tag, use it as is
	VERSION := $(EXACT_TAG)
endif

LDFLAGS := -ldflags "-X github.com/JammingBen/opencloud-skill-cli/internal/version.Version=$(VERSION)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/oc-cli cmd/oc-cli/*.go

.PHONY: release
release: build
	goreleaser release
	rm -rf dist

.PHONY: install
install:
	go install ./cmd/oc-cli

.PHONY: login
login:
	go run $(LDFLAGS) cmd/oc-cli/*.go login --server-url $(SERVER_URL) --insecure

.PHONY: logout
logout:
	go run $(LDFLAGS) cmd/oc-cli/*.go logout

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v -count=1 ./internal/client/

.PHONY: test-fuzz
test-fuzz:
	@for fuzz in ChunkSizes Offsets Filenames JSONBodies PathParams TUSOffsets HTTPMethods; do \
		echo "=== Fuzz$$fuzz ===" && \
		go test -run='^$$' -fuzz="^Fuzz$$fuzz$$" -fuzztime=5s ./internal/client/ || exit 1; \
	done

.PHONY: generate-skill-references
generate-skill-references:
	npx openapi-to-skills https://raw.githubusercontent.com/opencloud-eu/libre-graph-api/refs/heads/main/api/openapi-spec/v1.0.yaml -o ./output --name oc-libre-graph-api --exclude-paths /v1.0/education/users,/v1.0/education/users/{user-id},/v1.0/education/schools,/v1.0/education/schools/{school-id},/v1.0/education/schools/{school-id}/users,/v1.0/education/schools/{school-id}/users/$$ref,/v1.0/education/schools/{school-id}/users/{user-id}/$$ref,/v1.0/education/schools/{school-id}/classes,/v1.0/education/schools/{school-id}/classes/$$ref,/v1.0/education/schools/{school-id}/classes/{class-id}/$$ref,/v1.0/education/classes,/v1.0/education/classes/{class-id},/v1.0/education/classes/{class-id}/members,/v1.0/education/classes/{class-id}/members/$$ref,/v1.0/education/classes/{class-id}/members/{user-id}/$$ref,/v1.0/education/classes/{class-id}/teachers,/v1.0/education/classes/{class-id}/teachers/$$ref,/v1.0/education/classes/{class-id}/teachers/{user-id}/$$ref,/v1.0/me/drive/root,/v1.0/drives/{drive-id}/root && \
	rm -rf skills/opencloud-cli/references && \
	mkdir -p skills/opencloud-cli/references && \
	mv output/oc-libre-graph-api/references/operations skills/opencloud-cli/references && \
	mv output/oc-libre-graph-api/references/resources skills/opencloud-cli/references && \
	mv output/oc-libre-graph-api/references/schemas skills/opencloud-cli/references && rm -rf output