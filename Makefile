VERSION=0.2.4

check:
	go test -v .
	go test -race

lint:
	golangci-lint run ./...