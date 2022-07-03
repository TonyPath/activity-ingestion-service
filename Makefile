
test:
	go test ./... -coverprofile=coverage.out -count=1

coverage: test
	go tool cover -func coverage.out

report: coverage
	go tool cover -html=coverage.out -o cover.html


bump-deps:
	go get -u -v ./...
	go mod tidy