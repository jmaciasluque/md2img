.PHONY: build test vet bench install clean

VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X github.com/jmaciasluque/md2img.Version=$(VERSION)"

build:
	go build $(LDFLAGS) -o md2img ./cmd/md2img

test:
	go test -v -race -count=1 ./...

vet:
	go vet ./...

bench:
	go test -bench=. -benchmem -count=3 -timeout=300s ./...

install:
	go install $(LDFLAGS) ./cmd/md2img

clean:
	rm -f md2img
