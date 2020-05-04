.PHONY: build
build:
	GOOS=linux GOARCH=arm go build -o ./naskit-ui ./cmd/einkui/main.go