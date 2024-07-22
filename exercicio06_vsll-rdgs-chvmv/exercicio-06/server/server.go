package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	pb "exercicio-06/proto/exercicio-06/proto"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTimeServiceServer
}

func (s *server) Get(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	serverAddress := req.Host + ":2233"
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	request := req.Method + " " + req.Url + " HTTP/1.1\r\n" +
		"Host: " + req.Host + "\r\n" +
		"Connection: close\r\n" +
		"\r\n"

	_, err = conn.Write([]byte(request))
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[parts[0]] = strings.TrimSpace(parts[1])
		}
	}

	body := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		body += line
	}

	return &pb.HttpResponse{
		Status:  status,
		Headers: headers,
		Body:    body,
	}, nil
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	currentTime := time.Now().Format(time.RFC1123)
	fmt.Fprintln(w, currentTime)
}

func main() {
	lis, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTimeServiceServer(s, &server{})

	http.HandleFunc("/time", timeHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":2233", nil))
	}()

	fmt.Println("gRPC server running on port 12345")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
