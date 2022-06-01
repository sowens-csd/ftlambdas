.PHONY: vendor
test:
	go test -race ./...

build:
	go build -mod=vendor .

lambda-build:
	GOOS=linux go build -o /asset-output/main