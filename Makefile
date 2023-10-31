tests:
	go test -race -count=1 -coverprofile=coverage.out ./...

code-coverage:
	go tool cover -func=coverage.out

doc:
	godoc -http=:9090

lint:
	golangci-lint run

deprecations:
	grep "Deprecated: " -A 3 -R -n . | grep ".go"

benchmark:
	go test container/benchmark_test.go -bench=Container -benchmem

addlicense:
	addlicense -f LICENSE -ignore=vendor/\*\* -ignore=.github/\*\* .

addlicense-check:
	addlicense -f LICENSE -ignore=vendor/\*\* -ignore=.github/\*\* -check .
