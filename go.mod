module github.com/toransahu/grpc-eg-go

go 1.14

require (
	github.com/golang/mock v1.6.0
	github.com/lighttiger2505/sqls v0.0.1
	github.com/sourcegraph/jsonrpc2 v0.1.0
	golang.org/x/net v0.11.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230629202037-9506855d4529 // indirect
	google.golang.org/grpc v1.56.1
	google.golang.org/protobuf v1.31.0
)

replace github.com/lighttiger2505/sqls => ./../sqls
