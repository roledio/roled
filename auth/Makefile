VERSION := 0.20.0
COMMIT_HASH := $(shell git log -1 --pretty=format:%h || echo ${GIT_COMMIT_HASH})
BUILD_TIMESTAMP := $(shell date -u -Iseconds)
LDFLAGS := -X 'github.com/roledio/roled/auth/internal/models.AppVersion=${VERSION}'
LDFLAGS := ${LDFLAGS} -X 'github.com/roledio/roled/auth/internal/models.BuildTimestamp=${BUILD_TIMESTAMP}'
LDFLAGS := ${LDFLAGS} -X 'github.com/roledio/roled/auth/internal/models.CommitHash=${COMMIT_HASH}'

test-verbose:
	go test -v --cover ./...

test:
	go test -v -covermode=count -coverprofile=coverage.out $(shell go list ./... | egrep -v '/cmd|/configs|/constants|/entities|/errors|/mocks|/models|/views|/migrations|/databases|/types') -json > report.json

coverage:
	go tool cover -func=coverage.out

coverage-html:
	go tool cover -html=coverage.out

build:
	CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o ./tmp/app ./cmd/

.PHONY: version next-version bump-version

version: # Prints the current version
	@echo ${VERSION}

next-version: # Prints the next minor version without modifying the file
	@echo $(shell bash ./../.scripts/version.sh increment ${VERSION} 1)

bump-version: # Increments the minor version and updates this Makefile
	$(eval NEXT_VERSION := $(shell bash ./../.scripts/version.sh increment ${VERSION} 1))
	@echo "Bumping version from ${VERSION} to ${NEXT_VERSION}"
	sed -i -e "s~${VERSION}~${NEXT_VERSION}~" Makefile
