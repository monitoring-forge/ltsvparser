VERSION=0.2.3

check:
	go test -v .
	go test -race

lint:
	golangci-lint run ./...