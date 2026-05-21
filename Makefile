.PHONY: test
test:
	@go generate ./pkg/... # generate mocks; requires github.com/uber-go/mock
	@go test ./... -count=1 -v
