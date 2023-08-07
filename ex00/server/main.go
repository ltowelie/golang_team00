package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"transmitter/proto/message"
)

const (
	minMean        = -10.0
	maxMean        = 10.0 + math.SmallestNonzeroFloat64
	minStandardDev = 0.3
	maxStandardDev = 1.5 + math.SmallestNonzeroFloat64
)

type server struct {
	message.UnimplementedMessageServiceServer
}

func (s *server) StreamFrequency(mr *message.MessageRequest, ms message.MessageService_StreamFrequencyServer) error {

	uuidObj, mean, standardDeviation, err := calcParams()
	if err != nil {
		log.Printf("Error while generating parametrs: %s\n", err)
		return err
	}

	log.Printf("Connected new client. Parametrs are seesion id: %s, mean: %.4f, sd: %.4f\n", uuidObj.String(), mean, standardDeviation)

	messageResponse := &message.MessageResponse{}
	messageResponse.SessionId = uuidObj.String()

	for {
		frequency := rand.NormFloat64()*standardDeviation + mean
		messageResponse.Frequency = frequency
		messageResponse.Timestamp = time.Now().UnixMicro()
		err = ms.Send(messageResponse)
		if err != nil {
			log.Printf("Error while sending message to client with session id %s: %s\n", uuidObj.String(), err)
			return err
		}
	}
}

func calcParams() (*uuid.UUID, float64, float64, error) {
	uuidObj, err := uuid.NewRandom()
	if err != nil {
		return nil, 0, 0, err
	}

	mean := minMean + rand.Float64()*(maxMean-minMean)
	standardDeviation := minStandardDev + rand.Float64()*(maxStandardDev-minStandardDev)
	return &uuidObj, mean, standardDeviation, nil
}

func main() {

	host := os.Getenv("GRPCHOST")
	port := os.Getenv("GRPCPORT")
	addr := fmt.Sprintf("%s:%s", host, port)
	if addr == ":" {
		log.Fatal("No exported required env vars host or/and port")
	}
	log.Printf("Listening on %s", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error when create listener\n", err)
	}
	grpcServer := grpc.NewServer()

	message.RegisterMessageServiceServer(grpcServer, &server{})

	if err = grpcServer.Serve(listener); err != nil {
		log.Fatalf("Error when serve grpc server\n", err)
	}

}
