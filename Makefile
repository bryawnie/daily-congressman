BINARIES := retrieval chooser jobs api

.PHONY: build test vet fmt clean

build:
	@mkdir -p bin
	@for b in $(BINARIES); do go build -o bin/$$b ./cmd/$$b || exit 1; done

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w cmd internal

clean:
	rm -rf bin
