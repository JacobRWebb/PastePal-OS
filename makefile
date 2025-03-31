run:
	go mod tidy
	go run ./cmd/pastepal

build:
	go mod tidy
	go build -o pastepal.exe ./cmd/pastepal
