build:
	go build -o bin/yeahapi cmd/yeahapi/main.go

test:
	go test ./... -v
run:
	go run cmd/yeahapi/main.go
