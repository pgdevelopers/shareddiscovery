.PHONY: build


TEST_TARGETS=./

test:
	go test $(TEST_TARGETS) -coverprofile=coverage.txt -covermode=atomic --cover

lint:
	golangci-lint run --allow-parallel-runners --no-config --deadline=10m --enable=deadcode --enable=revive --enable=varcheck --enable=structcheck --enable=gocyclo --enable=errcheck --enable=gofmt --enable=goimports --enable=misspell --enable=unparam --enable=nakedret --enable=prealloc --enable=bodyclose --enable=gosec --enable=megacheck --exclude=G505 --exclude=G401

docgen:
	gopages -out docs -base "https://pgdevelopers.github.io/shareddiscovery" -brand-title "Lightyear Discovery Library"

mockgen:
	./scripts/mocks.sh
