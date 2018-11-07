# gRPCBench

go:generate protoc -I protobuf/ protobuf/engage.proto --go_out=plugins=grpc:protobuf

## execution

Run server:

    go run server/main.go

Run client:

    go run client/main.go
