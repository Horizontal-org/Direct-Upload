.PHONY: test prep build

test:
	go test -cover ./...

prep:
	go mod tidy
	go fmt ./...
	go build ./...
	go test ./...

build:
	go build github.com/horizontal-org/tus/

windows:
	CGO_ENABLED=0 GOOS=windows go build -a -installsuffix cgo -o tus.exe
