````markdown
# 📡 Go Observability Collector

A lightweight and pluggable **Go library** for sending structured **logs** and **metrics** to your own gRPC endpoint. Ideal for embedding into any Go microservice or infrastructure component.

---

## ✨ Features

✅ Unified library for Logs and Metrics  
✅ Buffered and batch-based transmission  
✅ Automatic flush on interval or size  
✅ Graceful shutdown with `SIGINT` / `SIGTERM`  
✅ Rich metadata (hostname, service name, OS, caller info, etc.)  
✅ Pluggable into any Go project

---

## 📦 Installation

```bash
go get github.com/yourusername/observability-collector
````

> Replace with your real module path or structure if you keep logs and metrics as separate packages like `logger` and `metriccollector`.

---

## 🔌 gRPC Backend Requirements

You’ll need a gRPC backend with two services:

### ✅ LogReceiver Service

```proto
service LogReceiver {
  rpc ReceiveLogs (LogBatch) returns (LogResponse);
}
```

### ✅ MetricReceiver Service

```proto
service MetricReceiver {
  rpc ReceiveMetrics (MetricBatch) returns (MetricResponse);
}
```

---

## 🛠️ Quickstart

### 1. Initialize Log Collector

```go
import "your_module_path/logger"

logCollector := logger.NewLogCollector("localhost:50051", "my-service", 10, 5*time.Second)

logCollector.Info("Application started")
logCollector.Error("Something failed: %v", err)
```

### 2. Initialize Metric Collector

```go
import "your_module_path/metriccollector"

metricCollector := metriccollector.NewMetricCollector("localhost:50052", "my-service", 10, 5*time.Second)

metricCollector.Collect("cpu.usage", metriccollector.GAUGE, 0.83, []string{"core:0"}, "percent")
```

---

## 📋 Metric Types

| Type           | Description                  |
| -------------- | ---------------------------- |
| `GAUGE`        | Value that can go up or down |
| `COUNT`        | Cumulative counter           |
| `HISTOGRAM`    | Distribution of values       |
| `RATE`         | Rate of change over time     |
| `SUMMARY`      | Percentiles                  |
| `METER`        | Events per time unit         |
| `DISTRIBUTION` | Custom distributions         |
| `SET`          | Unique values over time      |

---

## 🧼 Graceful Shutdown

Both collectors handle termination (`SIGINT`, `SIGTERM`) and flush remaining entries:

```bash
🛑 Shutdown detected! Flushing remaining logs/metrics...
🚀 LogCollector stopped.
🚀 MetricCollector stopped.
```

Or call manually:

```go
logCollector.Stop()
metricCollector.Stop()
```

---

## 📄 Example (Combined)

```go
package main

import (
	"time"
	"your_module_path/logger"
	"your_module_path/metriccollector"
)

func main() {
	logs := logger.NewLogCollector("localhost:50051", "combined-service", 10, 5*time.Second)
	metrics := metriccollector.NewMetricCollector("localhost:50052", "combined-service", 10, 5*time.Second)

	logs.Info("🚀 Starting service")
	metrics.Collect("requests.total", metriccollector.COUNT, 1, []string{"route:/health"}, "req")

	time.Sleep(10 * time.Second)

	logs.Stop()
	metrics.Stop()
}
```

---

---

## 📃 License

MIT © madhavan-21

---

## 👥 Contributors

Built by madhavan-21, inspired by modern observability platforms.

```
