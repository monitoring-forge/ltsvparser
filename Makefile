VERSION=0.2.7

check:
	go test -v .
	go test -race

lint:
	golangci-lint run ./...

bench:
	go test -bench '^BenchmarkEach' -benchmem -run=^$ ./...
	go test -bench '^BenchmarkByteToFloat' -benchmem -run=^$ ./...
