package anomalies

import (
	"fmt"
	"gorm.io/gorm"
	"math"
	"sync"
	"time"

	db "anomaliesToDB/db"
)

type AnomaliesDetector struct {
	SessionID                      string
	mu                             sync.Mutex
	Mean                           float64
	Sd                             float64
	CountRecords                   uint64
	CountAnomalies                 uint64
	sumForMean                     float64
	sumForSD                       float64
	anomalyCoefficient             float64
	initValues                     []float64
	freqCountToCalculateParameters uint64
	db                             *gorm.DB
}

func (a *AnomaliesDetector) InitCalculateMeanSD(freq float64) bool {

	a.mu.Lock()
	defer a.mu.Unlock()
	a.CountRecords++
	a.initValues = append(a.initValues, freq)

	if a.CountRecords == a.freqCountToCalculateParameters-1 {

		for _, freqForMeanSum := range a.initValues {
			a.sumForMean += freqForMeanSum
		}
		a.Mean = a.sumForMean / float64(a.CountRecords)

		for _, freqForSD := range a.initValues {
			a.sumForSD += math.Pow(freqForSD-a.Mean, 2)
		}
		a.Sd = math.Sqrt(a.sumForSD / float64(a.CountRecords))

		a.initValues = a.initValues[:0]
		return true

	}

	return false
}

func (a *AnomaliesDetector) ProcessNextFrequency(freq float64, timestamp int64) {

	a.mu.Lock()
	defer a.mu.Unlock()
	a.CountRecords++

	previousMean := a.Mean
	a.sumForMean += freq
	a.Mean = a.sumForMean / float64(a.CountRecords)

	a.sumForSD += math.Pow(freq-previousMean, 2)
	a.Sd = math.Sqrt(a.sumForSD / float64(a.CountRecords))

	calculatedFreqWithCoef := math.Abs(freq - a.Mean)
	if calculatedFreqWithCoef > a.anomalyCoefficient*a.Sd {
		fmt.Printf("Found anomaly %.4f\n", freq)

		anomaly := db.Anomaly{
			SessionID: a.SessionID,
			Frequency: freq,
			Timestamp: timestamp,
		}

		anomaly.CreateRecord(*a.db)

		a.CountAnomalies++
	}
}

func (a *AnomaliesDetector) DetectAnomalies(freqChan chan float64, stopChan chan struct{}) {

	go func(freqChan chan float64) {
		var freq float64
		var ok bool
		fmt.Printf("Compute parameters stage (need %d frequences to calculate).\n", a.freqCountToCalculateParameters)
		computed := false
		for !computed {
			freq, ok = <-freqChan
			if !ok {
				return
			}
			computed = a.InitCalculateMeanSD(freq)
		}
		fmt.Printf("Anomaly detection stage. Parameters - mean: %.4f sd: %.4f k*sd: %.4f\n",
			a.Mean,
			a.Sd,
			a.Sd*a.anomalyCoefficient)
		for {
			freq, ok = <-freqChan
			if !ok {
				return
			}
			now := time.Now().UnixMicro()
			a.ProcessNextFrequency(freq, now)
		}
	}(freqChan)

	go func() {
		ticker := time.Tick(time.Second * 5)
		for {
			<-ticker
			a.mu.Lock()
			fmt.Printf("Total records processed %d, anomalies found: %d\n",
				a.CountRecords,
				a.CountAnomalies)
			a.mu.Unlock()
			select {
			case <-stopChan:
				break
			default:
				continue
			}
		}
	}()
}

func NewAnomaliesDetector(anomalyCoef float64, db *gorm.DB, sessId string, freqCountToCalc uint64) *AnomaliesDetector {

	a := AnomaliesDetector{
		SessionID:                      sessId,
		Mean:                           0,
		Sd:                             0,
		CountRecords:                   0,
		CountAnomalies:                 0,
		sumForMean:                     0,
		sumForSD:                       0,
		anomalyCoefficient:             anomalyCoef,
		initValues:                     make([]float64, 0, freqCountToCalc),
		freqCountToCalculateParameters: freqCountToCalc,
		db:                             db,
	}
	return &a
}
