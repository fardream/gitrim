package svc

//go:generate protoc -I=. --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative svc.proto config.proto db.proto
