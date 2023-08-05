# protoc-gen-emmylua
protobuf protoc generate emmylua declare and enum definition

### build

```shell
go build -o protoc-gen-emmylua main.go
```
pls add protoc-gen-emmylua add to path environment

### gen emmylua
```shell
protoc --emmylua_out=./ test.proto
```