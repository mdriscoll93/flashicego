BINARY := flashicego

.PHONY: build test tidy fmt

build:
	go build -o $(BINARY) .

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
