package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"transmitter/proto/message"
)

func main() {

	host := os.Getenv("GRPCHOST")
	port := os.Getenv("GRPCPORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Error while creating tcp dial:%s\n", err)
	}
	defer conn.Close()

	client := message.NewMessageServiceClient(conn)
	ctx := context.Background()
	stream, err := client.StreamFrequency(ctx, &message.MessageRequest{})
	if err != nil {
		log.Fatalf("Client receiving stream failed: %v", err)
	}

	msg, err := stream.Recv()
	fmt.Printf("Received first message. Session id is %s timestamp %d. Frequenty is %f\n",
		msg.GetSessionId(),
		msg.GetTimestamp(),
		msg.GetFrequency())

	for {
		msg, err = stream.Recv()
		if err != nil {
			log.Fatalf("Error while receiving msg from server")
		}
		fmt.Printf("Timestamp: %d, frequenty is %f\n", msg.GetTimestamp(), msg.GetFrequency())
		// To generate frequency for anomalies detector
		//fmt.Printf("%f\n", msg.GetFrequency())
	}
}
