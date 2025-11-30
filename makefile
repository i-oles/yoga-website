run:
	go run cmd/yoga/main.go

lint:
	golangci-lint run ./...

test:
	go test -v ./...

build:
	go build -v -o bin/yoga cmd/yoga/main.go
