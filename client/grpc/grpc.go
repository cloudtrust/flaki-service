package main

import (
	"context"
	fb "github.com/JohanDroz/flaki-service/service/transport/flatbuffer/flaki"
	"github.com/go-kit/kit/log"
	"github.com/google/flatbuffers/go"
	"google.golang.org/grpc"
	"os"
)

const (
	address = "localhost:5555"
)

func main() {

	// Logger.
	var logger = log.NewLogfmtLogger(os.Stdout)
	{
		logger = log.With(logger, "time", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
		defer logger.Log("msg", "Goodbye")
	}
	logger = log.With(logger, "transport", "grpc")

	// Set up a connection to the server.
	var clienConn *grpc.ClientConn
	{
		var err error
		clienConn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithCodec(flatbuffers.FlatbuffersCodec{}))
		if err != nil {
			logger.Log("error", err)
		}
		defer clienConn.Close()
	}

	var flakiClient = fb.NewFlakiClient(clienConn)

	// Empty request
	var b = flatbuffers.NewBuilder(0)
	fb.EmptyRequestStart(b)
	b.Finish(fb.EmptyRequestEnd(b))

	// gRPC NextID
	var nextIDReply *fb.NextIDReply
	{
		var err error
		nextIDReply, err = flakiClient.NextID(context.Background(), b)
		if err != nil {
			logger.Log("error", err)
		}
		logger.Log("endpoint", "nextID", "id", nextIDReply.Id(), "error", nextIDReply.Error())
	}

	// gRPC NextValidID
	var nextValidIDReply *fb.NextValidIDReply
	{
		var err error
		nextValidIDReply, err = flakiClient.NextValidID(context.Background(), b)
		if err != nil {
			logger.Log("error", err)
		}
		logger.Log("endpoint", "nextValidID", "id", nextValidIDReply.Id())
	}
}
