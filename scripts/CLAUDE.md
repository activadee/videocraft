# Scripts Package - Python Integration Layer

## Overview
The `scripts` package provides Python-based integration tools for VideoCraft's AI transcription capabilities. It contains a dedicated Whisper daemon that enables efficient speech-to-text processing with optimized resource management.

## Architecture

```
scripts/
├── whisper_daemon.py      # Main Whisper transcription daemon
├── requirements.txt       # Python dependencies
└── CLAUDE.md             # This documentation
```

## Core Components

### whisper_daemon.py
**Location**: `scripts/whisper_daemon.py`

A persistent Python daemon that handles audio transcription requests via stdin/stdout communication. The daemon implements idle timeout for resource optimization and provides word-level timing for progressive subtitles.

#### Key Features

- **Persistent Process**: Long-running daemon that avoids model reloading overhead
- **Idle Timeout**: Automatic shutdown after 5 minutes of inactivity
- **JSON Communication**: Clean stdin/stdout protocol with Go services
- **Word-Level Timing**: Precise timestamps for progressive subtitle generation
- **Error Handling**: Comprehensive error reporting and recovery
- **Memory Management**: Automatic model unloading and garbage collection

#### Class Architecture

```python
class WhisperDaemon:
    def __init__(self, idle_timeout: int = 300, model_name: str = "base"):
        self.idle_timeout = idle_timeout
        self.model_name = model_name
        self.model = None
        self.device = self._get_optimal_device()
        self.last_activity = time.time()
        self.running = True
        self.shutdown_event = threading.Event()
```

#### Core Methods

##### `transcribe_audio(request: Dict[str, Any]) -> Dict[str, Any]`
Primary transcription method that processes audio URLs and returns detailed results.

**Input Parameters**:
```json
{
    "action": "transcribe",
    "url": "https://example.com/audio.wav",
    "language": "auto",
    "word_timestamps": true,
    "id": "request-123"
}
```

**Response Format**:
```json
{
    "success": true,
    "text": "Transcribed text content",
    "language": "en",
    "duration": 45.2,
    "segments": [
        {
            "start": 0.0,
            "end": 3.5,
            "text": "Hello world",
            "words": [
                {"word": "Hello", "start": 0.0, "end": 0.8},
                {"word": "world", "start": 1.2, "end": 1.8}
            ]
        }
    ],
    "word_timestamps": [
        {"word": "Hello", "start": 0.0, "end": 0.8},
        {"word": "world", "start": 1.2, "end": 1.8}
    ],
    "id": "request-123"
}
```

**Processing Flow**:
1. **Model Loading**: Lazy loads Whisper model on first request
2. **Audio Download**: Downloads audio to temporary file with SSL handling
3. **Transcription**: Runs Whisper with optimized parameters
4. **Word Extraction**: Extracts word-level timestamps from segments
5. **Response Generation**: Creates structured JSON response
6. **Cleanup**: Removes temporary files

##### `_get_optimal_device() -> str`
Determines the best available device for Whisper processing.

**Device Priority**:
1. **CUDA**: If available and compatible
2. **CPU**: Fallback for maximum compatibility

**Note**: MPS (Apple Silicon) is currently disabled due to sparse tensor issues with Whisper.

##### `_idle_checker() -> None`
Background thread that monitors activity and triggers automatic shutdown.

**Behavior**:
- Checks every 10 seconds for activity
- Triggers shutdown after 5 minutes of inactivity
- Helps preserve system resources

##### `handle_request(request: Dict[str, Any]) -> Dict[str, Any]`
Main request dispatcher that handles different action types.

**Supported Actions**:

1. **transcribe**: Main transcription functionality
2. **ping**: Health check (returns "pong")
3. **status**: Daemon status information
4. **shutdown**: Graceful shutdown command

**Status Response Example**:
```json
{
    "success": true,
    "model": "base",
    "device": "cpu",
    "model_loaded": true,
    "last_activity": 1640995200.0,
    "idle_timeout": 300,
    "id": "status-request"
}
```

## Go-Python Integration

### Communication Protocol

The daemon communicates with Go services through a simple stdin/stdout JSON protocol:

**Go Service -> Python Daemon**:
```go
type TranscriptionRequest struct {
    Action          string `json:"action"`
    URL             string `json:"url"`
    Language        string `json:"language,omitempty"`
    WordTimestamps  bool   `json:"word_timestamps"`
    ID              string `json:"id"`
}

// Send request
requestJSON, _ := json.Marshal(request)
daemon.stdin.Write(append(requestJSON, '\n'))

// Read response
responseJSON, _ := daemon.stdout.ReadLine()
var result TranscriptionResult
json.Unmarshal(responseJSON, &result)
```

**Python Daemon -> Go Service**:
```python
# Read request from stdin
line = sys.stdin.readline()
request = json.loads(line.strip())

# Process request
response = self.handle_request(request)

# Send response to stdout
print(json.dumps(response), flush=True)
```

### Lifecycle Management

#### Daemon Startup
```go
func (ts *transcriptionService) startDaemon() error {
    ts.cmd = exec.Command(ts.cfg.PythonPath, ts.cfg.WhisperDaemonPath,
        "--model", ts.cfg.WhisperModel,
        "--idle-timeout", "300")
    
    stdin, _ := ts.cmd.StdinPipe()
    stdout, _ := ts.cmd.StdoutPipe()
    stderr, _ := ts.cmd.StderrPipe()
    
    ts.stdin = stdin
    ts.stdout = bufio.NewReader(stdout)
    ts.stderr = bufio.NewReader(stderr)
    
    return ts.cmd.Start()
}
```

#### Health Monitoring
```go
func (ts *transcriptionService) healthCheck() error {
    request := map[string]interface{}{
        "action": "ping",
        "id":     "health-check",
    }
    
    response, err := ts.sendRequest(request)
    if err != nil || !response["success"].(bool) {
        return errors.New("daemon unhealthy")
    }
    
    return nil
}
```

#### Graceful Shutdown
```go
func (ts *transcriptionService) Shutdown() {
    if ts.cmd != nil && ts.cmd.Process != nil {
        // Send shutdown command
        shutdownCmd := map[string]interface{}{
            "action": "shutdown",
            "id":     "shutdown",
        }
        ts.sendRequest(shutdownCmd)
        
        // Wait for process to exit
        ts.cmd.Wait()
    }
}
```

## Progressive Subtitles Integration

### Word-Level Timing
The daemon provides precise word-level timestamps essential for progressive subtitle generation:

```python
# Extract word timestamps for progressive subtitles
word_timestamps_list = []
if "segments" in result:
    for segment in result["segments"]:
        if "words" in segment:
            word_timestamps_list.extend(segment["words"])

response = {
    "word_timestamps": word_timestamps_list,
    # ... other fields
}
```

### Go Integration
The Go subtitle service uses these timestamps to create progressive reveals:

```go
func (ss *subtitleService) createProgressiveEvents(words []models.WordTimestamp, sceneStart float64) []SubtitleEvent {
    var events []SubtitleEvent
    
    for _, word := range words {
        event := SubtitleEvent{
            StartTime: time.Duration((sceneStart + word.Start) * float64(time.Second)),
            EndTime:   time.Duration((sceneStart + word.End) * float64(time.Second)),
            Text:      word.Word,
            Type:      "progressive",
        }
        events = append(events, event)
    }
    
    return events
}
```

## Dependencies

### requirements.txt
```
openai-whisper>=20231117
torch>=2.0.0
torchaudio>=2.0.0
```

**Dependency Details**:

1. **openai-whisper**: Core Whisper AI model for speech recognition
2. **torch**: PyTorch framework required by Whisper
3. **torchaudio**: Audio processing utilities for PyTorch

### Installation
```bash
cd scripts
pip install -r requirements.txt
```

**System Requirements**:
- Python 3.8+
- 4GB+ RAM (for larger models)
- Internet connection for initial model download

## Configuration

### Command Line Arguments
```bash
python whisper_daemon.py \
    --model base \
    --idle-timeout 300 \
    --log-level INFO
```

**Available Models**:
- `tiny`: Fastest, least accurate (~39 MB)
- `base`: Good balance (~74 MB) - **Default**
- `small`: Better accuracy (~244 MB)
- `medium`: High accuracy (~769 MB)
- `large-v3`: Best accuracy (~1550 MB)

### Environment Integration
```go
type PythonConfig struct {
    Path              string `mapstructure:"path"`
    WhisperDaemonPath string `mapstructure:"whisper_daemon_path"`
    WhisperModel      string `mapstructure:"whisper_model"`
    WhisperDevice     string `mapstructure:"whisper_device"`
}
```

**Configuration Example**:
```yaml
python:
  path: "/usr/bin/python3"
  whisper_daemon_path: "./scripts/whisper_daemon.py"
  whisper_model: "base"
  whisper_device: "auto"
```

## Error Handling

### Daemon Error Responses
```python
# Transcription failure
{
    "success": false,
    "error": "Failed to download audio: 404 Not Found",
    "traceback": "...",
    "id": "request-123"
}

# JSON parsing error
{
    "success": false,
    "error": "Invalid JSON: Expecting ',' delimiter: line 1 column 45 (char 44)"
}
```

### Go Error Handling
```go
func (ts *transcriptionService) TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error) {
    response, err := ts.sendRequest(request)
    if err != nil {
        return nil, fmt.Errorf("daemon communication failed: %w", err)
    }
    
    if !response["success"].(bool) {
        errorMsg := response["error"].(string)
        return nil, fmt.Errorf("transcription failed: %s", errorMsg)
    }
    
    // Parse successful response
    result := parseTranscriptionResult(response)
    return result, nil
}
```

### Common Issues & Solutions

**1. Model Download Failures**
```bash
# Manual model download
python -c "import whisper; whisper.load_model('base')"
```

**2. Memory Issues**
```bash
# Use smaller model
python whisper_daemon.py --model tiny
```

**3. SSL Certificate Errors**
```python
# Daemon handles SSL issues automatically with:
ssl_context = ssl.create_default_context()
ssl_context.check_hostname = False
ssl_context.verify_mode = ssl.CERT_NONE
```

## Performance Optimization

### Model Caching
The daemon keeps the Whisper model loaded in memory between requests:

```python
def _load_model(self) -> None:
    """Load Whisper model if not already loaded"""
    if self.model is None:
        self.model = whisper.load_model(self.model_name, device=self.device)

def transcribe_audio(self, request: Dict[str, Any]) -> Dict[str, Any]:
    # Model stays loaded between requests
    self._load_model()  # No-op if already loaded
```

### Memory Management
```python
def _unload_model(self) -> None:
    """Unload model to free memory"""
    if self.model is not None:
        del self.model
        self.model = None
        
        # Force garbage collection
        if torch and torch.cuda.is_available():
            torch.cuda.empty_cache()
```

### Optimized Transcription Parameters
```python
result = self.model.transcribe(
    temp_path,
    language=None if language == "auto" else language,
    word_timestamps=True,
    verbose=False,        # Reduce output noise
    temperature=0,        # Deterministic results
    best_of=1,           # Single pass for speed
    beam_size=1          # Greedy decoding for speed
)
```

## Testing

### Manual Testing
```bash
# Start daemon
python scripts/whisper_daemon.py --model tiny

# Send test request
echo '{"action":"ping","id":"test"}' | python scripts/whisper_daemon.py

# Expected response
{"success": true, "message": "pong", "id": "test"}
```

### Integration Testing
```go
func TestWhisperDaemon_Integration(t *testing.T) {
    service := NewTranscriptionService(config)
    defer service.Shutdown()
    
    result, err := service.TranscribeAudio(context.Background(), "test-audio-url")
    
    assert.NoError(t, err)
    assert.NotEmpty(t, result.Text)
    assert.Greater(t, len(result.WordTimestamps), 0)
}
```

## Monitoring

### Daemon Status
```bash
# Check daemon status
echo '{"action":"status","id":"status"}' | python scripts/whisper_daemon.py
```

### Go Service Metrics
```go
type TranscriptionMetrics struct {
    RequestCount     int64         `json:"request_count"`
    SuccessCount     int64         `json:"success_count"`
    ErrorCount       int64         `json:"error_count"`
    AverageLatency   time.Duration `json:"average_latency"`
    DaemonUptime     time.Duration `json:"daemon_uptime"`
    ModelLoaded      bool          `json:"model_loaded"`
}
```

## Best Practices

### Resource Management
1. **Idle Timeout**: Use appropriate timeout for your workload
2. **Model Selection**: Choose smallest model that meets accuracy needs
3. **Memory Monitoring**: Monitor daemon memory usage
4. **Graceful Shutdown**: Always call Shutdown() on service stop

### Error Handling
1. **Retry Logic**: Implement retry for transient failures
2. **Timeout Handling**: Set reasonable timeouts for requests
3. **Daemon Recovery**: Restart daemon on critical failures
4. **Logging**: Log all transcription requests and errors

### Security Considerations
1. **SSL Handling**: Daemon handles SSL certificate issues
2. **Temporary Files**: Automatic cleanup of downloaded audio
3. **Input Validation**: Validate URLs before sending to daemon
4. **Resource Limits**: Monitor and limit concurrent requests

## Development Guidelines

### Adding New Features
1. **Protocol Extension**: Add new action types to handle_request()
2. **Response Format**: Maintain consistent JSON response structure
3. **Error Handling**: Comprehensive error reporting for debugging
4. **Testing**: Add unit tests for new functionality

### Debugging
```python
# Enable debug logging (avoid in production)
logging.getLogger().setLevel(logging.DEBUG)

# Test specific requests
python scripts/whisper_daemon.py --log-level DEBUG
```

### Contributing
1. Follow PEP 8 style guidelines
2. Add type hints for all functions
3. Include comprehensive docstrings
4. Test with multiple audio formats
5. Validate memory usage and performance