package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "exercicio-06/proto/exercicio-06/proto"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:12345", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewTimeServiceClient(conn)

	req := &pb.HttpRequest{
		Method: "GET",
		Url:    "/time",
		Host:   "localhost",
	}

	numRequests := 1
	var totalElapsed time.Duration

	for i := 0; i < numRequests; i++ {
		startTime := time.Now()

		res, err := c.Get(context.Background(), req) // _ = res
		if err != nil {
			fmt.Printf("Error on request %d: %v\n", i+1, err)
			continue
		}

		elapsed := time.Since(startTime)
		totalElapsed += elapsed

		fmt.Printf("Server Time %d: %s\n", i+1, res.Body)
		fmt.Printf("RPC request %d took %s\n", i+1, elapsed)
	}

	fmt.Printf("Total time taken for %d requests: %s\n", numRequests, totalElapsed)
	fmt.Printf("Average time per request: %s\n", totalElapsed/time.Duration(numRequests))
}
