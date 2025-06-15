
````markdown
# 📦 Go Log & Metrics Collector

A lightweight Go library to collect and ship logs and metrics to a gRPC-based endpoint. This library is ideal for observability pipelines and can be integrated easily into any Go service.

---

## 🚀 Features

- Buffered logging with flush thresholds
- Automatic periodic flush
- Structured logs with metadata (timestamp, service, hostname, file, line, function)
- Graceful shutdown handling (logs flushed on exit)
- Supports custom log levels: `DEBUG`, `INFO`, `WARN`, `ERROR`
- Sends logs to your gRPC server using Protobuf

---

## 📦 Installation

```bash
go get github.com/yourusername/logger
````

> Replace with your actual module path if hosted privately.

---

## 🛠️ Usage

### 1. Import the package

```go
import "your_module_path/logger"
```

### 2. Initialize the log collector

```go
lc := logger.NewLogCollector("localhost:50051", "my-service", 10, 5*time.Second)
```

* `localhost:50051`: gRPC server address
* `"my-service"`: your service name
* `10`: flush logs after 10 messages
* `5s`: flush logs every 5 seconds

### 3. Send logs

```go
lc.Info("Starting the service")
lc.Debug("Debugging value: %v", value)
lc.Warn("Something looks suspicious")
lc.Error("Something went wrong: %v", err)
```

---

## 🧼 Graceful Shutdown

The collector handles `SIGINT` and `SIGTERM`, ensuring logs are flushed before exit:

```bash
CTRL+C
🛑 Shutdown detected! Flushing remaining logs...
🚀 LogCollector stopped.
```

You can also stop it manually:

```go
lc.Stop()
```

---

## 🧪 Protobuf

This library expects a gRPC server with the following service:

```proto
service LogReceiver {
  rpc ReceiveLogs (LogBatch) returns (LogResponse);
}

message Log {
  string message = 1;
  string level = 2;
  google.protobuf.Timestamp timestamp = 3;
  string hostname = 4;
  string service = 5;
  string file = 6;
  int32 line = 7;
  string function = 8;
}

message LogBatch {
  repeated Log logs = 1;
}

message LogResponse {
  string status = 1;
}
```

---

## 📎 Example

```go
package main

import (
	"time"
	"your_module_path/logger"
)

func main() {
	lc := logger.NewLogCollector("localhost:50051", "demo-app", 5, 10*time.Second)

	lc.Info("Application started")
	lc.Debug("Loading configuration")

	// Simulate work
	time.Sleep(15 * time.Second)

	lc.Stop()
}
```

---

## 📄 License

MIT © \madhavan-21

```

Let me know if you want to include a **metrics sending** feature in the README too — or want to split it into `log_collector` and `metrics_collector` sections.
```





```protobuf command


 protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative protobuf/metrics.proto
```
