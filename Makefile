tests: tests-concurrency
	go test -race -count=1 -coverprofile=coverage.out ./...

tests-concurrency:
	go test -race -run concurrency -tags concurrency -count=1 ./...

code-coverage:
	go tool cover -func=coverage.out

doc:
	godoc -http=:9090

lint:
	golangci-lint run
