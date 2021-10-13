.PHONY: build


TEST_TARGETS=./

test:
	go test $(TEST_TARGETS) -coverprofile=coverage.txt -covermode=atomic --cover

docgen:
	gopages -out docs -base "https://pgdevelopers.github.io/shareddiscovery" -brand-title "Lightyear Discovery Library"

mockgen:
	./scripts/mocks.sh
