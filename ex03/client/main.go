package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	anomalies_detector "anomalies/anomalies_detector"
	anomalies_db "anomalies/db"
	"anomalies/proto/message"
)

const (
	freqCountToCalculateParameters = 100
)

func main() {

	var anomalyCoefficient float64
	flag.Float64Var(&anomalyCoefficient, "k", 0, "STD anomaly coefficient")
	flag.Parse()

	if len(flag.Args()) != 0 || anomalyCoefficient == 0 {
		flag.Usage()
		os.Exit(1)
	}

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

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DBHOST"),
		os.Getenv("DBUSER"),
		os.Getenv("DBPASSWORD"),
		os.Getenv("DBNAME"),
		os.Getenv("DBPORT"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %s", err)
	}
	anomalies_db.Migrate(*db)

	msg, err := stream.Recv()
	fmt.Printf("Received first message. Session id is %s timestamp %d. Frequenty is %f\n",
		msg.GetSessionId(),
		msg.GetTimestamp(),
		msg.GetFrequency())

	freqChan := make(chan *message.MessageResponse)
	stop := make(chan struct{})
	defer close(freqChan)
	anomaliesDetector := anomalies_detector.NewAnomaliesDetector(
		anomalyCoefficient,
		db,
		msg.GetSessionId(),
		freqCountToCalculateParameters)
	anomaliesDetector.DetectAnomalies(freqChan, stop)

	for {
		msg, err = stream.Recv()
		if err != nil {
			stop <- struct{}{}
			log.Fatalf("Error while receiving msg from server")
		}
		freqChan <- msg
	}

}
