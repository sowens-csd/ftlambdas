.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/passwordlessAuthorizer lambdas/passwordlessAuthorizer/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/jwtAuthorizer lambdas/jwtAuthorizer/main.go

	env GOOS=linux go build -ldflags="-s -w" -o bin/query lambdas/query/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/delete lambdas/delete/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/authCleanup lambdas/authCleanup/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
