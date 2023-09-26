tests:
	go test -race -count=1 -coverprofile=coverage.out ./...

code-coverage:
	go tool cover -func=coverage.out

doc:
	godoc -http=:9090

lint:
	golangci-lint run
