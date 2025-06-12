# Config Package - Configuration Management

## Overview
The `internal/config` package provides centralized configuration management for VideoCraft using Viper. It handles configuration loading from multiple sources (files, environment variables, defaults) and provides strongly-typed configuration structures for all application components.

## Architecture

```text
internal/config/
├── config.go         # Main configuration structures and loading logic
└── CLAUDE.md         # This documentation
```

## Configuration Structure

### Main Configuration

```go
type Config struct {
    Server        ServerConfig        `mapstructure:"server"`
    FFmpeg        FFmpegConfig        `mapstructure:"ffmpeg"`
    Transcription TranscriptionConfig `mapstructure:"transcription"`
    Subtitles     SubtitlesConfig     `mapstructure:"subtitles"`
    Storage       StorageConfig       `mapstructure:"storage"`
    Job           JobConfig           `mapstructure:"job"`
    Log           LogConfig           `mapstructure:"log"`
    Security      SecurityConfig      `mapstructure:"security"`
}
```

### Server Configuration

```go
type ServerConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}

func (s ServerConfig) Address() string {
    return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
```

**Purpose**: Configures the HTTP server binding and network settings.

**Default Values**:
- `host`: "0.0.0.0" (bind to all interfaces)
- `port`: 3002

**Environment Variables**:
- `VIDEOCRAFT_SERVER_HOST`
- `VIDEOCRAFT_SERVER_PORT`

**Usage Example**:
```yaml
server:
  host: "0.0.0.0"
  port: 3002
```

### FFmpeg Configuration

```go
type FFmpegConfig struct {
    BinaryPath string        `mapstructure:"binary_path"`
    Timeout    time.Duration `mapstructure:"timeout"`
    Quality    int           `mapstructure:"quality"`
    Preset     string        `mapstructure:"preset"`
}
```

**Purpose**: Configures FFmpeg binary location and encoding parameters.

**Parameters**:
- `binary_path`: Path to FFmpeg executable
- `timeout`: Maximum processing time for FFmpeg operations
- `quality`: CRF value for video quality (0-51, lower is better)
- `preset`: FFmpeg encoding preset (ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow)

**Default Values**:
- `binary_path`: "ffmpeg" (expects FFmpeg in PATH)
- `timeout`: "1h" (1 hour maximum processing time)
- `quality`: 23 (good balance of quality and file size)
- `preset`: "medium" (good balance of speed and compression)

**Environment Variables**:
- `VIDEOCRAFT_FFMPEG_BINARY_PATH`
- `VIDEOCRAFT_FFMPEG_TIMEOUT`
- `VIDEOCRAFT_FFMPEG_QUALITY`
- `VIDEOCRAFT_FFMPEG_PRESET`

**Usage Example**:
```yaml
ffmpeg:
  binary_path: "/usr/local/bin/ffmpeg"
  timeout: "2h"
  quality: 20
  preset: "fast"
```

### Transcription Configuration

```go
type TranscriptionConfig struct {
    Enabled    bool             `mapstructure:"enabled"`
    Daemon     DaemonConfig     `mapstructure:"daemon"`
    Python     PythonConfig     `mapstructure:"python"`
    Processing ProcessingConfig `mapstructure:"processing"`
}

type DaemonConfig struct {
    Enabled             bool          `mapstructure:"enabled"`
    IdleTimeout         time.Duration `mapstructure:"idle_timeout"`
    StartupTimeout      time.Duration `mapstructure:"startup_timeout"`
    RestartMaxAttempts  int           `mapstructure:"restart_max_attempts"`
}

type PythonConfig struct {
    Path       string `mapstructure:"path"`
    ScriptPath string `mapstructure:"script_path"`
    Model      string `mapstructure:"model"`
    Language   string `mapstructure:"language"`
    Device     string `mapstructure:"device"`
}

type ProcessingConfig struct {
    Workers int           `mapstructure:"workers"`
    Timeout time.Duration `mapstructure:"timeout"`
}
```

**Purpose**: Configures the Python Whisper daemon for speech recognition.

**Daemon Configuration**:
- `enabled`: Enable/disable daemon mode
- `idle_timeout`: Time before daemon auto-shuts down
- `startup_timeout`: Maximum time to wait for daemon startup
- `restart_max_attempts`: Maximum restart attempts on failure

**Python Configuration**:
- `path`: Python executable path
- `script_path`: Directory containing Python scripts
- `model`: Whisper model size (tiny, base, small, medium, large)
- `language`: Default language ("auto" for auto-detection)
- `device`: Processing device ("cpu", "cuda", "auto")

**Processing Configuration**:
- `workers`: Number of concurrent transcription workers
- `timeout`: Maximum time per transcription request

**Default Values**:
```yaml
transcription:
  enabled: true
  daemon:
    enabled: true
    idle_timeout: "300s"    # 5 minutes
    startup_timeout: "30s"
    restart_max_attempts: 3
  python:
    path: "python3"
    script_path: "./scripts"
    model: "base"
    language: "auto"
    device: "auto"
  processing:
    workers: 2
    timeout: "60s"
```

**Model Selection Guide**:
- `tiny`: ~39 MB, fastest, least accurate
- `base`: ~74 MB, good balance (recommended)
- `small`: ~244 MB, better accuracy
- `medium`: ~769 MB, high accuracy
- `large`: ~1550 MB, best accuracy

### Subtitles Configuration

```go
type SubtitlesConfig struct {
    Enabled    bool        `mapstructure:"enabled"`
    Style      string      `mapstructure:"style"`
    FontFamily string      `mapstructure:"font_family"`
    FontSize   int         `mapstructure:"font_size"`
    Position   string      `mapstructure:"position"`
    Colors     ColorConfig `mapstructure:"colors"`
}

type ColorConfig struct {
    Word    string `mapstructure:"word"`
    Outline string `mapstructure:"outline"`
}
```

**Purpose**: Configures subtitle generation and styling.

**Parameters**:
- `enabled`: Enable/disable subtitle generation
- `style`: Subtitle style ("progressive" or "classic")
- `font_family`: Font family for subtitles
- `font_size`: Font size in points
- `position`: Subtitle position on screen
- `colors.word`: Text color (hex format)
- `colors.outline`: Outline color (hex format)

**Style Types**:
- `progressive`: Character-by-character reveal synchronized with speech
- `classic`: Traditional word-level or sentence-level subtitles

**Position Options**:
- `center-bottom`: Centered at bottom (default)
- `center-top`: Centered at top
- `left-bottom`: Left-aligned at bottom
- `right-bottom`: Right-aligned at bottom

**Default Values**:
```yaml
subtitles:
  enabled: true
  style: "progressive"
  font_family: "Arial"
  font_size: 24
  position: "center-bottom"
  colors:
    word: "#FFFFFF"
    outline: "#000000"
```

### Storage Configuration

```go
type StorageConfig struct {
    OutputDir       string        `mapstructure:"output_dir"`
    TempDir         string        `mapstructure:"temp_dir"`
    MaxFileSize     int64         `mapstructure:"max_file_size"`
    CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
    RetentionDays   int           `mapstructure:"retention_days"`
}
```

**Purpose**: Configures file storage and cleanup policies.

**Parameters**:
- `output_dir`: Directory for generated videos
- `temp_dir`: Directory for temporary files
- `max_file_size`: Maximum file size in bytes
- `cleanup_interval`: How often to run cleanup
- `retention_days`: Days to retain generated files

**Default Values**:
```yaml
storage:
  output_dir: "./generated_videos"
  temp_dir: "./temp"
  max_file_size: 1073741824  # 1GB
  cleanup_interval: "1h"
  retention_days: 7
```

**File Size Limits**:
- 1GB default limit prevents excessive storage usage
- Configurable based on available disk space
- Enforced during file upload and processing

### Job Configuration

```go
type JobConfig struct {
    Workers             int           `mapstructure:"workers"`
    QueueSize           int           `mapstructure:"queue_size"`
    MaxConcurrent       int           `mapstructure:"max_concurrent"`
    StatusCheckInterval time.Duration `mapstructure:"status_check_interval"`
}
```

**Purpose**: Configures job processing and concurrency limits.

**Parameters**:
- `workers`: Number of worker goroutines for job processing
- `queue_size`: Maximum number of queued jobs
- `max_concurrent`: Maximum concurrent job processing
- `status_check_interval`: Frequency of job status updates

**Default Values**:
```yaml
job:
  workers: 4
  queue_size: 100
  max_concurrent: 10
  status_check_interval: "5s"
```

**Concurrency Considerations**:
- Higher worker count increases parallelism but uses more resources
- Queue size prevents memory exhaustion during high load
- Max concurrent limits system resource usage

### Logging Configuration

```go
type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}
```

**Purpose**: Configures application logging.

**Parameters**:
- `level`: Log level (debug, info, warn, error)
- `format`: Log format (text, json)

**Default Values**:
```yaml
log:
  level: "debug"
  format: "text"
```

**Log Levels**:
- `debug`: Verbose logging for development
- `info`: General information messages
- `warn`: Warning messages
- `error`: Error messages only

### Security Configuration

```go
type SecurityConfig struct {
    APIKey     string `mapstructure:"api_key"`
    RateLimit  int    `mapstructure:"rate_limit"`
    EnableAuth bool   `mapstructure:"enable_auth"`
}
```

**Purpose**: Configures security and access control.

**Parameters**:
- `api_key`: API key for authentication (when enabled)
- `rate_limit`: Requests per minute per client
- `enable_auth`: Enable/disable API key authentication

**Default Values**:
```yaml
security:
  rate_limit: 100
  enable_auth: false
  # api_key: "your-secret-key-here"
```

**Security Notes**:
- API key should be set via environment variable in production
- Rate limiting prevents abuse and resource exhaustion
- Authentication can be disabled for development

## Configuration Loading

### Load Function

```go
func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("/etc/videocraft/")

    // Set defaults
    setDefaults()

    // Environment variables
    viper.AutomaticEnv()
    viper.SetEnvPrefix("VIDEOCRAFT")

    // Read config file
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

### Configuration Sources

**Priority Order** (highest to lowest):
1. **Environment Variables**: `VIDEOCRAFT_SECTION_KEY`
2. **Configuration File**: `config.yaml`
3. **Default Values**: Built-in defaults

### Configuration File Locations

The loader searches for configuration files in:
1. Current directory (`./config.yaml`)
2. Config subdirectory (`./config/config.yaml`)
3. System directory (`/etc/videocraft/config.yaml`)

### Environment Variable Mapping

Environment variables use the prefix `VIDEOCRAFT_` followed by the configuration path:

```bash
# Server configuration
export VIDEOCRAFT_SERVER_HOST="localhost"
export VIDEOCRAFT_SERVER_PORT="8080"

# FFmpeg configuration
export VIDEOCRAFT_FFMPEG_BINARY_PATH="/usr/local/bin/ffmpeg"
export VIDEOCRAFT_FFMPEG_QUALITY="20"

# Transcription configuration
export VIDEOCRAFT_TRANSCRIPTION_PYTHON_MODEL="small"
export VIDEOCRAFT_TRANSCRIPTION_DAEMON_IDLE_TIMEOUT="600s"

# Security configuration
export VIDEOCRAFT_SECURITY_API_KEY="your-secret-key"
export VIDEOCRAFT_SECURITY_ENABLE_AUTH="true"
```

## Usage Examples

### Basic Configuration File

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 3002

ffmpeg:
  binary_path: "ffmpeg"
  quality: 23
  preset: "medium"

transcription:
  enabled: true
  python:
    model: "base"
    device: "cpu"

subtitles:
  enabled: true
  style: "progressive"
  font_size: 24

storage:
  output_dir: "./videos"
  temp_dir: "./tmp"

security:
  enable_auth: false
```

### Production Configuration

```yaml
# config/production.yaml
server:
  host: "0.0.0.0"
  port: 80

ffmpeg:
  timeout: "2h"
  quality: 20
  preset: "fast"

transcription:
  python:
    model: "small"  # Better accuracy for production
    device: "cuda"  # Use GPU if available
  processing:
    workers: 8
    timeout: "120s"

job:
  workers: 16
  max_concurrent: 20

storage:
  output_dir: "/var/videocraft/output"
  temp_dir: "/var/videocraft/temp"
  retention_days: 30

security:
  enable_auth: true
  rate_limit: 500

log:
  level: "info"
  format: "json"
```

### Development Configuration

```yaml
# config/development.yaml
server:
  port: 3000

ffmpeg:
  quality: 30  # Faster encoding for development
  preset: "fast"

transcription:
  python:
    model: "tiny"  # Fastest model for development

subtitles:
  font_size: 20

storage:
  retention_days: 1  # Clean up quickly in development

security:
  enable_auth: false  # Disabled for development

log:
  level: "debug"
  format: "text"
```

## Configuration Validation

### Runtime Validation

```go
func (c *Config) Validate() error {
    // Server validation
    if c.Server.Port <= 0 || c.Server.Port > 65535 {
        return errors.New("invalid server port")
    }

    // FFmpeg validation
    if c.FFmpeg.Quality < 0 || c.FFmpeg.Quality > 51 {
        return errors.New("ffmpeg quality must be between 0 and 51")
    }

    validPresets := map[string]bool{
        "ultrafast": true, "superfast": true, "veryfast": true,
        "faster": true, "fast": true, "medium": true,
        "slow": true, "slower": true, "veryslow": true,
    }
    if !validPresets[c.FFmpeg.Preset] {
        return errors.New("invalid ffmpeg preset")
    }

    // Transcription validation
    if c.Transcription.Enabled {
        validModels := map[string]bool{
            "tiny": true, "base": true, "small": true,
            "medium": true, "large": true,
        }
        if !validModels[c.Transcription.Python.Model] {
            return errors.New("invalid whisper model")
        }
    }

    // Storage validation
    if c.Storage.MaxFileSize <= 0 {
        return errors.New("max file size must be positive")
    }

    // Job validation
    if c.Job.Workers <= 0 {
        return errors.New("job workers must be positive")
    }

    return nil
}
```

### Environment-Specific Validation

```go
func (c *Config) ValidateForEnvironment(env string) error {
    switch env {
    case "production":
        if c.Security.EnableAuth && c.Security.APIKey == "" {
            return errors.New("API key required in production")
        }
        if c.Log.Level == "debug" {
            return errors.New("debug logging not recommended in production")
        }
    case "development":
        // Development-specific validations
    }
    return c.Validate()
}
```

## Advanced Configuration

### Dynamic Configuration Reloading

```go
type ConfigManager struct {
    config *Config
    mu     sync.RWMutex
    
    // Configuration change callbacks
    callbacks []func(*Config)
}

func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        callbacks: make([]func(*Config), 0),
    }
}

func (cm *ConfigManager) GetConfig() *Config {
    cm.mu.RLock()
    defer cm.mu.RUnlock()
    return cm.config
}

func (cm *ConfigManager) ReloadConfig() error {
    newConfig, err := Load()
    if err != nil {
        return err
    }

    if err := newConfig.Validate(); err != nil {
        return err
    }

    cm.mu.Lock()
    oldConfig := cm.config
    cm.config = newConfig
    cm.mu.Unlock()

    // Notify callbacks of configuration change
    for _, callback := range cm.callbacks {
        callback(newConfig)
    }

    return nil
}

func (cm *ConfigManager) OnConfigChange(callback func(*Config)) {
    cm.callbacks = append(cm.callbacks, callback)
}
```

### Configuration Encryption

```go
func LoadEncryptedConfig(key []byte) (*Config, error) {
    // Read encrypted config file
    encryptedData, err := ioutil.ReadFile("config.encrypted")
    if err != nil {
        return nil, err
    }

    // Decrypt configuration
    decryptedData, err := decrypt(encryptedData, key)
    if err != nil {
        return nil, err
    }

    // Parse decrypted YAML
    var config Config
    if err := yaml.Unmarshal(decryptedData, &config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

## Best Practices

### Configuration Management

1. **Environment Variables for Secrets**: Use environment variables for sensitive data
2. **Layered Configuration**: Use file + environment variable overrides
3. **Validation**: Always validate configuration at startup
4. **Documentation**: Document all configuration options
5. **Defaults**: Provide sensible defaults for all options

### Security Considerations

1. **Secret Management**: Never commit secrets to version control
2. **Environment Separation**: Use different configs for different environments
3. **Minimal Permissions**: Grant minimal required permissions
4. **Encryption**: Encrypt sensitive configuration data
5. **Auditing**: Log configuration changes

### Performance Optimization

1. **Resource Limits**: Set appropriate resource limits
2. **Concurrency**: Tune concurrency based on hardware
3. **Caching**: Cache configuration values when appropriate
4. **Monitoring**: Monitor configuration-dependent metrics
5. **Profiling**: Profile with different configuration settings

### Development Workflow

1. **Local Overrides**: Use local config files for development
2. **Environment Consistency**: Keep environments as similar as possible
3. **Configuration Testing**: Test with different configuration combinations
4. **Documentation**: Keep configuration documentation up to date
5. **Validation**: Validate configuration in CI/CD pipelines

## Configuration Examples

### Docker Configuration

```yaml
# docker-compose.yml environment
environment:
  VIDEOCRAFT_SERVER_HOST: "0.0.0.0"
  VIDEOCRAFT_SERVER_PORT: "80"
  VIDEOCRAFT_FFMPEG_BINARY_PATH: "/usr/bin/ffmpeg"
  VIDEOCRAFT_STORAGE_OUTPUT_DIR: "/app/output"
  VIDEOCRAFT_STORAGE_TEMP_DIR: "/app/temp"
  VIDEOCRAFT_TRANSCRIPTION_PYTHON_PATH: "/usr/bin/python3"
  VIDEOCRAFT_SECURITY_API_KEY: "${API_KEY}"
  VIDEOCRAFT_SECURITY_ENABLE_AUTH: "true"
```

### Kubernetes Configuration

```yaml
# ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: videocraft-config
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 8080
    
    ffmpeg:
      binary_path: "/usr/bin/ffmpeg"
      timeout: "2h"
    
    storage:
      output_dir: "/data/output"
      temp_dir: "/data/temp"
    
    transcription:
      python:
        model: "base"
        device: "cpu"

---
# Secret for sensitive data
apiVersion: v1
kind: Secret
metadata:
  name: videocraft-secrets
type: Opaque
data:
  api-key: <base64-encoded-api-key>
```

### CI/CD Configuration

```yaml
# .github/workflows/config-validation.yml
name: Configuration Validation
on: [push, pull_request]

jobs:
  validate-config:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Validate default configuration
        run: |
          go run cmd/validate-config/main.go --config config/config.yaml
      
      - name: Validate production configuration
        run: |
          go run cmd/validate-config/main.go --config config/production.yaml --env production
```

## Troubleshooting

### Common Configuration Issues

**1. Configuration File Not Found**
```bash
# Check search paths
ls -la config.yaml
ls -la config/config.yaml
ls -la /etc/videocraft/config.yaml
```

**2. Environment Variable Override Issues**
```bash
# Check environment variables
env | grep VIDEOCRAFT
```

**3. Permission Issues**
```bash
# Check file permissions
ls -la config/
chmod 600 config/config.yaml  # Secure sensitive configs
```

**4. YAML Syntax Errors**
```bash
# Validate YAML syntax
python3 -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

**5. Validation Failures**
```bash
# Run configuration validation
go run cmd/videocraft/main.go --validate-config
```

### Debugging Configuration Loading

```go
func debugConfigLoading() {
    // Enable Viper debugging
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    
    // Show what Viper is doing
    fmt.Printf("Config file used: %s\n", viper.ConfigFileUsed())
    fmt.Printf("All settings: %+v\n", viper.AllSettings())
    
    // Show environment variable mappings
    for _, key := range viper.AllKeys() {
        envKey := "VIDEOCRAFT_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
        fmt.Printf("%s -> %s = %v\n", envKey, key, viper.Get(key))
    }
}
```