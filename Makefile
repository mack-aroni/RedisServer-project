run: build
	@./bin/fs --listenAddr :5001

build: 
	@go build -o bin/fs .

test:
	@go test -v ./...