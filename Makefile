.PHONY: clean schema

build:
	@go build -buildmode=pie -o bin/bfkata main.go

build-m1:
	@GOOS=darwin GOARCH=arm64 go build -o bin/bf-macos-arm64 main.go

schema:
	protoc --go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:.  \
    	api/api.proto

clean:
	@rm -rf bin/bf
	@rm -rf bin/bf-macos-arm64


test: build
	@bin/bf test
