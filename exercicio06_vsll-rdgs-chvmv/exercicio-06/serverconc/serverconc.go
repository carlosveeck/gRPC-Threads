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
	"google.golang.org/grpc/reflection"
)

// Implementação do servidor que contém o método Get
type server struct {
	pb.UnimplementedTimeServiceServer
}

// Método Get processa a solicitação HTTP manualmente e retorna a resposta
func (s *server) Get(ctx context.Context, req *pb.HttpRequest) (*pb.HttpResponse, error) {
	responseChan := make(chan *pb.HttpResponse)
	errorChan := make(chan error)

	// Inicia uma goroutine para processar a solicitação
	go func() {
		serverAddr := req.Host + ":2233"
		conn, err := net.Dial("tcp", serverAddr)
		if err != nil {
			errorChan <- err
			return
		}
		defer conn.Close()

		request := fmt.Sprintf("%s %s HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", req.Method, req.Url, req.Host)
		_, err = conn.Write([]byte(request))
		if err != nil {
			errorChan <- err
			return
		}

		reader := bufio.NewReader(conn)
		status, err := reader.ReadString('\n')
		if err != nil {
			errorChan <- err
			return
		}

		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				errorChan <- err
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				headers[parts[0]] = parts[1]
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

		responseChan <- &pb.HttpResponse{
			Status:  status,
			Headers: headers,
			Body:    body,
		}
	}()

	select {
	case res := <-responseChan:
		return res, nil
	case err := <-errorChan:
		return nil, err
	}
}

func timeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, time.Now().Format(time.RFC1123))
}

func main() {
	lis, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterTimeServiceServer(s, &server{})
	reflection.Register(s)

	http.HandleFunc("/time", timeHandler)
	go http.ListenAndServe(":2233", nil) // Inicia o servidor HTTP em uma goroutine separada

	fmt.Println("gRPC server listening on port 12345...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
