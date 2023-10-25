package main

import (
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"

	"github.com/fardream/gitrim/cmd"
)

var printextop = prototext.MarshalOptions{
	Multiline: true,
	Indent:    "  ",
}

func PrintProtoText(m proto.Message) string {
	return string(cmd.GetOrPanic(printextop.Marshal(m)))
}
