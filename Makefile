build-api:
	go build -o bin/yeahapi cmd/yeahapi/main.go

build-ui:
	go build -o bin/yeahui cmd/yeahui/main.go

run-api:
	go run cmd/yeahapi/main.go

run-ui:
	go run cmd/yeahui/main.go

test:
	go test ./... -v
