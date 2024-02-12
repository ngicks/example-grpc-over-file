# example-grpc-over-file

## Environment

Assume you have latest Go installed.

install protobuf-compiler. see official guide: https://grpc.io/docs/protoc-installation/

If you are on Debian like system

```shell
sudo apt-get update && sudo apt install -y protobuf-compiler
protoc --version  # Ensure compiler version is 3+
```

install Go plugins

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

make sure `${GOPATH}/bin` is exported.

```
export PATH="$PATH:$(go env GOPATH)/bin"
```

If you update any of `.proto` file, run `go generate`

```
go generate ./...
```
