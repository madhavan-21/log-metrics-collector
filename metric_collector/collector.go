package metriccollector

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	pb "metrics-collector/protobuf"
	// Ensure this matches your actual protobuf package path

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricType represents the type of metric
type MetricType string

const (
	GAUGE        MetricType = "gauge"
	COUNT        MetricType = "count"
	HISTOGRAM    MetricType = "histogram"
	RATE         MetricType = "rate"
	SUMMARY      MetricType = "summary"
	METER        MetricType = "meter"
	DISTRIBUTION MetricType = "distribution"
	SET          MetricType = "set"
)

// Metric represents a collected metric
type Metric struct {
	Name        string     `json:"name"`
	Type        MetricType `json:"type"`
	Value       float64    `json:"value"`
	Timestamp   time.Time  `json:"timestamp"`
	Tags        []string   `json:"tags"`
	ProjectName string     `json:"project_name"`
	Hostname    string     `json:"hostname"`
	OS          string     `json:"os"`
	UniqueID    string     `json:"unique_id"`
	Unit        string     `json:"unit"`
}

// MetricCollector handles metric collection and sending
type MetricCollector struct {
	buffer      []Metric
	mu          sync.Mutex
	flushSize   int
	interval    time.Duration
	ticker      *time.Ticker
	client      pb.MetricReceiverClient
	wg          sync.WaitGroup
	stopChan    chan struct{}
	projectName string
	hostname    string
	osName      string
	uniqueID    string
}

// generateUniqueID creates a SHA-256 hash of (Project Name + Hostname)
func generateUniqueID(projectName, hostname string) string {
	data := projectName + hostname
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// NewMetricCollector initializes a MetricCollector
func NewMetricCollector(grpcAddress string, projectName string, flushSize int, interval time.Duration) *MetricCollector {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("❌ Failed to get hostname: %v", err)
	}

	osName := runtime.GOOS
	uniqueID := generateUniqueID(projectName, hostname)

	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("❌ Failed to connect to gRPC server: %v", err)
	}
	client := pb.NewMetricReceiverClient(conn) // Corrected to MetricReceiverClient

	mc := &MetricCollector{
		buffer:      make([]Metric, 0, flushSize),
		flushSize:   flushSize,
		interval:    interval,
		ticker:      time.NewTicker(interval),
		client:      client,
		stopChan:    make(chan struct{}),
		projectName: projectName,
		hostname:    hostname,
		osName:      osName,
		uniqueID:    uniqueID,
	}

	// Start background worker
	mc.wg.Add(1)
	go mc.startWorker()

	// Handle graceful shutdown
	go mc.handleShutdown()
	fmt.Println("✅ MetricCollector initialized")
	return mc
}

// Collect adds a metric to the buffer
func (mc *MetricCollector) Collect(name string, metricType MetricType, value float64, tags []string, unit string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metric := Metric{
		Name:        name,
		Type:        metricType,
		Value:       value,
		Timestamp:   time.Now(),
		Tags:        tags,
		ProjectName: mc.projectName,
		Hostname:    mc.hostname,
		OS:          mc.osName,
		UniqueID:    mc.uniqueID,
		Unit:        unit,
	}

	mc.buffer = append(mc.buffer, metric)

	if len(mc.buffer) >= mc.flushSize {
		mc.flush()
	}
}

// startWorker periodically flushes metrics
func (mc *MetricCollector) startWorker() {
	defer mc.wg.Done()
	for {
		select {
		case <-mc.ticker.C:
			mc.mu.Lock()
			if len(mc.buffer) > 0 {
				mc.flush()
			}
			mc.mu.Unlock()
		case <-mc.stopChan:
			mc.mu.Lock()
			if len(mc.buffer) > 0 {
				mc.flush()
			}
			mc.mu.Unlock()
			return
		}
	}
}

// flush sends metrics to the server
func (mc *MetricCollector) flush() {
	if len(mc.buffer) == 0 {
		return
	}

	protoMetrics := &pb.MetricBatch{}
	for _, m := range mc.buffer {
		metric := &pb.Metric{
			Name:        m.Name,
			Type:        string(m.Type),
			Value:       m.Value,
			Timestamp:   timestamppb.New(m.Timestamp),
			Tags:        m.Tags,
			ProjectName: m.ProjectName,
			Hostname:    m.Hostname,
			Os:          m.OS,
			UniqueId:    m.UniqueID,
			Unit:        m.Unit,
		}

		protoMetrics.Metrics = append(protoMetrics.Metrics, metric)
	}

	err := sendToServer(mc.client, protoMetrics)
	if err != nil {
		log.Println("❌ Failed to send metrics:", err)
		return
	}

	mc.buffer = make([]Metric, 0, mc.flushSize)
	mc.ticker.Reset(mc.interval)

	fmt.Printf("✅ Sent %d metrics, buffer cleared.\n", len(protoMetrics.Metrics))
}

// Stop stops the collector and flushes remaining metrics
func (mc *MetricCollector) Stop() {
	close(mc.stopChan)
	mc.wg.Wait()
	mc.ticker.Stop()
	fmt.Println("🚀 MetricCollector stopped.")
}

// handleShutdown ensures graceful shutdown
func (mc *MetricCollector) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n🛑 Shutdown detected! Flushing remaining metrics...")
	mc.Stop()
	os.Exit(0)
}

// sendToServer sends metrics to the gRPC server
func sendToServer(client pb.MetricReceiverClient, data *pb.MetricBatch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.ReceiveMetrics(ctx, data)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %v", err)
	}

	fmt.Println("✅ Server response:", resp.Status)
	return nil
}
