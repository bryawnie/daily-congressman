BINARIES := retrieval chooser jobs api

.PHONY: build test vet fmt clean webapp-check webapp-build

build:
	@mkdir -p bin
	@for b in $(BINARIES); do go build -o bin/$$b ./cmd/$$b || exit 1; done

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w cmd internal

webapp-check:
	cd webapp && npm run check

webapp-build:
	cd webapp && npm run build

clean:
	rm -rf bin webapp/build
