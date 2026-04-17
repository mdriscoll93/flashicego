BINARY := flashicego

.PHONY: build test integration-test tidy fmt

build:
	go build -o $(BINARY) .

test:
	go test ./...

integration-test:
	go test ./integration -v

tidy:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
