# Performance Guide

This document provides comprehensive guidance for optimizing VideoCraft performance, monitoring system resources, and troubleshooting performance issues.

## =Ê Performance Overview

VideoCraft is designed for high-performance video generation with several optimization strategies:

- **Concurrent Processing**: Multi-worker job processing with configurable parallelism
- **Daemon Architecture**: Long-running Whisper process for efficient transcription
- **Caching Strategy**: Intelligent caching for transcription results and metadata
- **Resource Management**: Configurable resource limits and cleanup policies
- **Streaming Operations**: Memory-efficient file handling and processing

## ™ Configuration Tuning

### Core Performance Settings

```yaml
# config.yaml - Performance optimized configuration
job:
  workers: 8                    # Number of worker goroutines
  queue_size: 200              # Job queue capacity
  max_concurrent: 16           # Maximum concurrent jobs
  status_check_interval: "2s"  # Status update frequency

ffmpeg:
  timeout: "2h"               # Processing timeout
  quality: 23                 # CRF value (lower = better quality, slower)
  preset: "medium"            # Encoding preset (fast/medium/slow)

transcription:
  daemon:
    enabled: true             # Use persistent daemon
    idle_timeout: "600s"      # Keep daemon alive longer
    startup_timeout: "45s"    # Allow more startup time
  processing:
    workers: 4                # Transcription workers
    timeout: "120s"           # Per-request timeout
  python:
    model: "base"             # Model size vs accuracy tradeoff
```

### Environment Variables for Production

```bash
# Performance Configuration
export VIDEOCRAFT_JOB_WORKERS="16"
export VIDEOCRAFT_JOB_MAX_CONCURRENT="32"
export VIDEOCRAFT_JOB_QUEUE_SIZE="500"

# FFmpeg Optimization
export VIDEOCRAFT_FFMPEG_PRESET="fast"
export VIDEOCRAFT_FFMPEG_QUALITY="25"
export VIDEOCRAFT_FFMPEG_TIMEOUT="3h"

# Transcription Performance
export VIDEOCRAFT_TRANSCRIPTION_DAEMON_ENABLED="true"
export VIDEOCRAFT_TRANSCRIPTION_DAEMON_IDLE_TIMEOUT="900s"
export VIDEOCRAFT_TRANSCRIPTION_PROCESSING_WORKERS="8"
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="small"

# Storage Optimization
export VIDEOCRAFT_STORAGE_CLEANUP_INTERVAL="30m"
export VIDEOCRAFT_STORAGE_RETENTION_DAYS="3"
```

## <¯ Performance Optimization

### CPU Optimization

#### Multi-Core Utilization
```go
// Configure for available CPU cores
runtime.GOMAXPROCS(runtime.NumCPU())

// Job processing parallelism
func optimizeForCPU(cfg *config.Config) {
    cpuCores := runtime.NumCPU()
    
    // Workers should match CPU cores
    cfg.Job.Workers = cpuCores
    
    // Allow more concurrent jobs on multi-core systems
    cfg.Job.MaxConcurrent = cpuCores * 2
    
    // Transcription workers based on available cores
    cfg.Transcription.Processing.Workers = cpuCores / 2
}
```

#### FFmpeg CPU Settings
```yaml
ffmpeg:
  # CPU-optimized presets
  preset: "fast"        # Good balance of speed/quality
  # preset: "veryfast"  # Maximum speed
  # preset: "medium"    # Better quality, slower
  
  # Quality settings for performance
  quality: 25           # Faster encoding (vs 23 default)
  
  # Additional FFmpeg arguments for CPU optimization
  extra_args:
    - "-threads"
    - "0"              # Use all available CPU threads
    - "-preset"
    - "fast"
```

### Memory Optimization

#### Memory Management
```go
// Memory-efficient job processing
type JobProcessor struct {
    maxMemoryMB int64
    currentJobs map[string]*Job
    memoryUsage int64
}

func (jp *JobProcessor) CanAcceptJob(estimatedMemoryMB int64) bool {
    return jp.memoryUsage + estimatedMemoryMB <= jp.maxMemoryMB
}

// Garbage collection optimization
func optimizeGC() {
    // Adjust GC target percentage for high-memory scenarios
    debug.SetGCPercent(50) // More frequent GC for video processing
    
    // Force GC after heavy operations
    runtime.GC()
}
```

#### Memory Configuration
```yaml
# Memory-optimized settings
job:
  max_concurrent: 8      # Limit concurrent jobs based on available RAM
  
storage:
  max_file_size: 2147483648  # 2GB max file size
  cleanup_interval: "15m"    # More frequent cleanup
  
transcription:
  python:
    model: "tiny"       # Use smaller model if memory constrained
    # model: "base"     # 74MB model (recommended)
    # model: "small"    # 244MB model
    # model: "medium"   # 769MB model
```

### I/O Optimization

#### Disk Performance
```yaml
storage:
  output_dir: "/fast-ssd/videocraft/output"  # Use SSD storage
  temp_dir: "/fast-ssd/videocraft/temp"      # Temporary files on SSD
  cleanup_interval: "10m"                    # Frequent cleanup
  retention_days: 1                          # Aggressive cleanup for development
```

#### Network Optimization
```go
// HTTP client optimization for file downloads
func optimizedHTTPClient() *http.Client {
    return &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
            IdleConnTimeout:     90 * time.Second,
            DisableCompression:  false,  // Enable compression
            
            // Connection pooling
            DialContext: (&net.Dialer{
                Timeout:   10 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
        },
    }
}
```

## =È Performance Monitoring

### System Metrics

#### Built-in Health Checks
```bash
# Basic health check
curl http://localhost:3002/health

# Detailed system information
curl http://localhost:3002/health/detailed
```

#### Response Example
```json
{
  "status": "healthy",
  "uptime": "2h30m45s",
  "system": {
    "go_version": "go1.21.5",
    "goroutines": 25,
    "memory": {
      "allocated": 15728640,
      "total_alloc": 157286400,
      "sys": 71303192,
      "heap_alloc": 15728640,
      "heap_sys": 67108864,
      "gc_cycles": 12
    }
  }
}
```

### Performance Profiling

#### CPU Profiling
```bash
# Enable CPU profiling
go build -o videocraft cmd/server/main.go
./videocraft -cpuprofile=cpu.prof

# Analyze CPU profile
go tool pprof cpu.prof
(pprof) top10
(pprof) web
```

#### Memory Profiling
```bash
# Memory profile via HTTP endpoint
go tool pprof http://localhost:3002/debug/pprof/heap

# Memory allocation profile
go tool pprof http://localhost:3002/debug/pprof/allocs

# Goroutine profile
go tool pprof http://localhost:3002/debug/pprof/goroutine
```

#### Benchmark Tests
```bash
# Run performance benchmarks
make benchmark

# Specific benchmark tests
go test -bench=BenchmarkVideoService ./internal/services/
go test -bench=BenchmarkSubtitleGeneration ./internal/services/
go test -bench=BenchmarkFFmpegProcessing ./internal/services/
```

### Metrics Collection

#### Prometheus Metrics
```go
// Performance metrics
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "videocraft_request_duration_seconds",
            Help: "Request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint", "status"},
    )
    
    jobProcessingTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "videocraft_job_processing_seconds",
            Help: "Job processing time in seconds",
            Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1200},
        },
        []string{"job_type", "status"},
    )
    
    activeJobs = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "videocraft_active_jobs",
            Help: "Number of currently active jobs",
        },
        []string{"status"},
    )
)
```

#### Custom Metrics Endpoint
```bash
# Metrics endpoint
curl http://localhost:3002/metrics

# Example metrics output
# HELP videocraft_request_duration_seconds Request duration in seconds
# TYPE videocraft_request_duration_seconds histogram
videocraft_request_duration_seconds_bucket{method="POST",endpoint="/generate-video",status="202",le="0.1"} 45
videocraft_request_duration_seconds_bucket{method="POST",endpoint="/generate-video",status="202",le="0.25"} 120
```

## =' Performance Tuning Guide

### Whisper Model Selection

| Model | Size | Speed | Accuracy | Use Case |
|-------|------|-------|----------|----------|
| `tiny` | 39 MB | Fastest | Basic | Development/Testing |
| `base` | 74 MB | Fast | Good | Production Default |
| `small` | 244 MB | Medium | Better | High-Accuracy Needs |
| `medium` | 769 MB | Slow | High | Maximum Accuracy |
| `large` | 1550 MB | Slowest | Highest | Specialized Use Cases |

### FFmpeg Preset Selection

| Preset | Speed | Quality | CPU Usage | Use Case |
|--------|-------|---------|-----------|----------|
| `ultrafast` | Fastest | Lowest | Low | Real-time/Streaming |
| `veryfast` | Very Fast | Low | Medium | High-volume Processing |
| `fast` | Fast | Good | Medium | Balanced Production |
| `medium` | Medium | Better | High | Quality Production |
| `slow` | Slow | High | Very High | Archive/Distribution |

### Concurrent Job Limits

```go
// Calculate optimal concurrency based on system resources
func calculateOptimalConcurrency() (workers, maxConcurrent int) {
    cpuCores := runtime.NumCPU()
    memoryGB := getAvailableMemoryGB()
    
    // Base calculation on CPU cores
    workers = cpuCores
    
    // Adjust for memory constraints
    // Assume each job uses ~500MB average
    memoryBasedLimit := int(memoryGB * 2) // Allow 2 jobs per GB
    
    if memoryBasedLimit < workers {
        workers = memoryBasedLimit
    }
    
    // Maximum concurrent jobs (including queued)
    maxConcurrent = workers * 3
    
    // Ensure minimum viable configuration
    if workers < 2 {
        workers = 2
    }
    if maxConcurrent < 4 {
        maxConcurrent = 4
    }
    
    return workers, maxConcurrent
}
```

## =€ Scaling Strategies

### Horizontal Scaling

#### Load Balancer Configuration
```nginx
# nginx.conf - Load balancing multiple instances
upstream videocraft_backend {
    server 127.0.0.1:3002;
    server 127.0.0.1:3003;
    server 127.0.0.1:3004;
    server 127.0.0.1:3005;
    
    # Load balancing method
    least_conn;  # Route to server with fewest active connections
}

server {
    listen 80;
    server_name api.yourdomain.com;
    
    location / {
        proxy_pass http://videocraft_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        
        # Timeouts for video processing
        proxy_connect_timeout 60s;
        proxy_send_timeout 300s;
        proxy_read_timeout 300s;
    }
}
```

#### Docker Swarm Scaling
```yaml
# docker-compose.yml for swarm mode
version: '3.8'
services:
  videocraft:
    image: videocraft:latest
    deploy:
      replicas: 4
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
```

### Vertical Scaling

#### Resource Allocation
```yaml
# High-performance single instance configuration
job:
  workers: 32               # Maximum workers for high-end server
  queue_size: 1000         # Large queue for burst capacity
  max_concurrent: 64       # High concurrency limit

ffmpeg:
  preset: "fast"           # Balanced speed/quality
  quality: 23              # Good quality

transcription:
  processing:
    workers: 16            # Many transcription workers
  python:
    model: "small"         # Larger model for better accuracy
```

## =Ê Performance Benchmarks

### Typical Performance Metrics

#### Video Generation Performance
| Video Length | Audio Files | Processing Time | Memory Usage |
|--------------|-------------|-----------------|--------------|
| 30 seconds | 1 file | 15-45 seconds | 200-500 MB |
| 2 minutes | 3 files | 1-3 minutes | 400-800 MB |
| 5 minutes | 5 files | 3-8 minutes | 600-1200 MB |
| 10 minutes | 8 files | 6-15 minutes | 1-2 GB |

#### Transcription Performance (Whisper Models)
| Model | Audio Length | Processing Time | Memory Usage |
|-------|--------------|-----------------|--------------|
| tiny | 1 minute | 5-10 seconds | 200 MB |
| base | 1 minute | 10-20 seconds | 300 MB |
| small | 1 minute | 20-40 seconds | 500 MB |
| medium | 1 minute | 40-80 seconds | 1 GB |

### Performance Testing

#### Load Testing Script
```bash
#!/bin/bash
# load-test.sh - Basic load testing

API_BASE="http://localhost:3002/api/v1"
CONCURRENT_REQUESTS=10
TOTAL_REQUESTS=100

# Function to make video generation request
generate_video() {
    curl -s -X POST "$API_BASE/generate-video" \
        -H "Content-Type: application/json" \
        -d '{
            "scenes": [{
                "id": "test",
                "elements": [{
                    "type": "audio",
                    "src": "https://example.com/test.mp3"
                }]
            }]
        }' > /dev/null
}

# Run concurrent requests
echo "Starting load test..."
start_time=$(date +%s)

for i in $(seq 1 $TOTAL_REQUESTS); do
    generate_video &
    
    # Limit concurrent requests
    if (( i % CONCURRENT_REQUESTS == 0 )); then
        wait
    fi
done

wait
end_time=$(date +%s)
duration=$((end_time - start_time))

echo "Load test completed in ${duration} seconds"
echo "Average: $((TOTAL_REQUESTS / duration)) requests/second"
```

## = Performance Troubleshooting

### Common Performance Issues

#### High Memory Usage
```bash
# Check memory usage
free -h
ps aux | grep videocraft | awk '{print $6}' | head -1

# Monitor Go memory
curl -s http://localhost:3002/health/detailed | jq '.system.memory'

# Solutions:
# 1. Reduce concurrent jobs
# 2. Use smaller Whisper model
# 3. Increase cleanup frequency
# 4. Add more RAM
```

#### High CPU Usage
```bash
# Check CPU usage
top -p $(pgrep videocraft)
htop

# Monitor goroutines
curl -s http://localhost:3002/health/detailed | jq '.system.goroutines'

# Solutions:
# 1. Reduce worker count
# 2. Use faster FFmpeg preset
# 3. Limit concurrent processing
# 4. Add CPU cores
```

#### Slow Processing
```bash
# Check processing bottlenecks
curl -s http://localhost:3002/health/detailed

# Monitor job queue
curl -s http://localhost:3002/api/v1/jobs | jq '.jobs[] | select(.status=="pending") | length'

# Solutions:
# 1. Increase worker count
# 2. Optimize FFmpeg settings
# 3. Use faster storage (SSD)
# 4. Check network latency for downloads
```

### Performance Optimization Checklist

- [ ] **CPU Optimization**
  - [ ] Set optimal worker count based on CPU cores
  - [ ] Use appropriate FFmpeg preset
  - [ ] Configure GOMAXPROCS correctly
  - [ ] Monitor goroutine count

- [ ] **Memory Optimization**
  - [ ] Choose appropriate Whisper model
  - [ ] Set reasonable concurrent job limits
  - [ ] Configure aggressive cleanup policies
  - [ ] Monitor heap allocation

- [ ] **I/O Optimization**
  - [ ] Use SSD storage for temp files
  - [ ] Optimize network settings
  - [ ] Configure appropriate timeouts
  - [ ] Enable HTTP connection pooling

- [ ] **Monitoring Setup**
  - [ ] Enable health check endpoints
  - [ ] Set up metrics collection
  - [ ] Configure alerting for resource usage
  - [ ] Implement performance logging

## =Ú Additional Resources

- [Go Performance Guide](https://github.com/golang/go/wiki/Performance)
- [FFmpeg Performance Optimization](https://trac.ffmpeg.org/wiki/EncodingForStreamingSites)
- [System Resource Monitoring](../troubleshooting/overview.md)
- [Docker Performance Best Practices](../deployment/docker.md#performance-optimization)

This performance guide provides comprehensive strategies for optimizing VideoCraft across various deployment scenarios and workloads.