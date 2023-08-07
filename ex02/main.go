package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	anomalies_detector "anomaliesToDB/anomalies_detector"
	anomalies_db "anomaliesToDB/db"
)

const (
	freqCountToCalculateParameters = 100
)

func main() {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DBHOST"),
		os.Getenv("DBUSER"),
		os.Getenv("DBPASSWORD"),
		os.Getenv("DBNAME"),
		os.Getenv("DBPORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	anomalies_db.Migrate(*db)

	var anomalyCoefficient float64
	flag.Float64Var(&anomalyCoefficient, "k", 0, "STD anomaly coefficient")
	flag.Parse()

	if len(flag.Args()) != 0 || anomalyCoefficient == 0 {
		flag.Usage()
		os.Exit(1)
	}

	freqChan := make(chan float64)
	stop := make(chan struct{})
	defer close(freqChan)
	defer close(stop)
	anomaliesDetector := anomalies_detector.NewAnomaliesDetector(
		anomalyCoefficient,
		db,
		"033026e6-69b8-471e-b8fe-a6aaebbb03f9",
		freqCountToCalculateParameters)
	anomaliesDetector.DetectAnomalies(freqChan, stop)

	var readValue float64

	for {
		_, err = fmt.Fscan(os.Stdin, &readValue)
		if err == io.EOF {
			stop <- struct{}{}
			break
		}
		if err != nil {
			stop <- struct{}{}
			log.Fatalf("Error reading value from stdin: %s", err)
		}
		freqChan <- readValue
	}
}
