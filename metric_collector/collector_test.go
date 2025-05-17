package metriccollector_test

import (
	"fmt"
	metriccollector "metrics-collector/metric_collector" // Update with the actual package path
	"testing"
	"time"
)

func TestMetricCollector(t *testing.T) {
	// grpcServer, grpcAddress, mockServer := startMockServer(t)
	// defer grpcServer.Stop()
	fmt.Println("ada")
	collector := metriccollector.NewMetricCollector("localhost:50051", "TestProject", 8, 15*time.Second)
	//defer collector.Stop()

	collector.Collect("cpu_usage", metriccollector.RATE, 78, []string{"env:dev"}, "bytes")
	collector.Collect("memory_usage", metriccollector.GAUGE, 70.2, []string{"env:dev"}, "%")
	for i := 0; i < 10; i++ {
		collector.Collect("custom_metics", metriccollector.GAUGE, float64(i), []string{"env:dev"}, "us")
	}

	//time.Sleep(450 * time.Second)
	//t.SetTimeout(60 * time.Second) // Increase the timeout
	// Wait for automatic flush
	select {}
	// assert.Equal(t, 2, len(mockServer.receivedMetrics), "Expected 2 metrics to be sent")
	// assert.Equal(t, "cpu_usage", mockServer.receivedMetrics[0].Name)
	// assert.Equal(t, "memory_usage", mockServer.receivedMetrics[1].Name)
}
