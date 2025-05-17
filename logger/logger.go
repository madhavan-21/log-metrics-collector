package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	pb "minion/metrics-collector/log_protobuf"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LogLevel represents log severity levels
type LogLevel string

const (
	DEBUG LogLevel = "debug"
	INFO  LogLevel = "info"
	WARN  LogLevel = "warn"
	ERROR LogLevel = "error"
)

// LogMessage represents a structured log entry
type LogMessage struct {
	Message   string   `json:"message"`
	Level     string   `json:"level"`
	Timestamp string   `json:"timestamp"`
	Hostname  string   `json:"hostname"`
	Service   string   `json:"service"`
	Source    string   `json:"source"`
	Tags      []string `json:"tags"`
	File      string   `json:"file"`
	Line      int      `json:"line"`
	Function  string   `json:"function"`
}

type LogCollector struct {
	buffer    []LogMessage
	mu        sync.Mutex
	flushSize int
	interval  time.Duration
	ticker    *time.Ticker
	client    pb.LogReceiverClient
	wg        sync.WaitGroup
	stopChan  chan struct{}
	hostname  string
	service   string
}

// Get the current timestamp
func getTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// Get the caller's file, line number, and function name
func getCallerInfo() (string, int, string) {
	pc, file, line, _ := runtime.Caller(3)
	fn := runtime.FuncForPC(pc).Name()
	return filepath.Base(file), line, fn
}

// NewLogCollector initializes a LogCollector
func NewLogCollector(grpcAddress, service string, flushSize int, interval time.Duration) *LogCollector {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("❌ Failed to get hostname: %v", err)
	}

	conn, err := grpc.Dial(grpcAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("❌ Failed to connect to gRPC server: %v", err)
	}
	client := pb.NewLogReceiverClient(conn)

	lc := &LogCollector{
		buffer:    make([]LogMessage, 0, flushSize),
		flushSize: flushSize,
		interval:  interval,
		ticker:    time.NewTicker(interval),
		client:    client,
		stopChan:  make(chan struct{}),
		hostname:  hostname,
		service:   service,
	}

	lc.wg.Add(1)
	go lc.startWorker()
	go lc.handleShutdown()
	fmt.Println("✅ LogCollector initialized")
	return lc
}

func (lc *LogCollector) log(level LogLevel, format string, args ...interface{}) {
	file, line, function := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logEntry := LogMessage{
		Message:   message,
		Level:     string(level),
		Timestamp: getTimestamp(),
		Hostname:  lc.hostname,
		Service:   lc.service,
		File:      file,
		Line:      line,
		Function:  function,
	}

	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.buffer = append(lc.buffer, logEntry)

	if len(lc.buffer) >= lc.flushSize {
		lc.flush()
	}
}

func (lc *LogCollector) flush() {
	if len(lc.buffer) == 0 {
		return
	}

	protoLogs := &pb.LogBatch{}
	for _, log := range lc.buffer {
		protoLog := &pb.Log{
			Message:   log.Message,
			Level:     log.Level,
			Timestamp: timestamppb.Now(),
			Hostname:  log.Hostname,
			Service:   log.Service,
			File:      log.File,
			Line:      int32(log.Line),
			Function:  log.Function,
		}
		protoLogs.Logs = append(protoLogs.Logs, protoLog)
	}

	err := lc.sendToServer(protoLogs)
	if err != nil {
		log.Println("❌ Failed to send logs:", err)
		return
	}

	lc.buffer = lc.buffer[:0]
	lc.ticker.Reset(lc.interval)
	fmt.Printf("✅ Sent %d logs, buffer cleared.\n", len(protoLogs.Logs))
}

func (lc *LogCollector) sendToServer(data *pb.LogBatch) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := lc.client.ReceiveLogs(ctx, data)
	if err != nil {
		return fmt.Errorf("gRPC request failed: %v", err)
	}

	fmt.Println("✅ Server response:", resp.Status)
	return nil
}

func (lc *LogCollector) startWorker() {
	defer lc.wg.Done()
	for {
		select {
		case <-lc.ticker.C:
			lc.mu.Lock()
			if len(lc.buffer) > 0 {
				lc.flush()
			}
			lc.mu.Unlock()
		case <-lc.stopChan:
			lc.mu.Lock()
			if len(lc.buffer) > 0 {
				lc.flush()
			}
			lc.mu.Unlock()
			return
		}
	}
}

func (lc *LogCollector) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Shutdown detected! Flushing remaining logs...")
	lc.Stop()
	os.Exit(0)
}

func (lc *LogCollector) Stop() {
	close(lc.stopChan)
	lc.wg.Wait()
	lc.ticker.Stop()
	fmt.Println("🚀 LogCollector stopped.")
}

func (lc *LogCollector) Debug(format string, args ...interface{}) { lc.log(DEBUG, format, args...) }
func (lc *LogCollector) Info(format string, args ...interface{})  { lc.log(INFO, format, args...) }
func (lc *LogCollector) Warn(format string, args ...interface{})  { lc.log(WARN, format, args...) }
func (lc *LogCollector) Error(format string, args ...interface{}) { lc.log(ERROR, format, args...) }
