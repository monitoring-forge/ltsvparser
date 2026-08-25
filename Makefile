VERSION=0.2.5

check:
	go test -v .
	go test -race

lint:
	golangci-lint run ./...