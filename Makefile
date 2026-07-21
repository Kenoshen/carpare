.PHONY: run build clean tidy

run:
	go run .

build:
	go build -o bin/carpare .

clean:
	rm -rf bin carpare

tidy:
	go mod tidy
