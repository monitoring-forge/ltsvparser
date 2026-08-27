VERSION=0.2.6

check:
	go test -v .
	go test -race

lint:
	golangci-lint run ./...

bench:
	go test -bench=. -benchmem -run=^$$