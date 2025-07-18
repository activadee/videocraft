# � Configuration Overview

VideoCraft provides flexible configuration through environment variables, YAML files, and runtime settings. This document covers all configuration options and best practices for different deployment scenarios.

## <� Configuration Hierarchy

VideoCraft loads configuration in the following order (higher priority overrides lower):

1. **Command Line Arguments** (highest priority)
2. **Environment Variables** 
3. **YAML Configuration Files**
4. **Default Values** (lowest priority)

```mermaid
graph TB
    CLI[Command Line Args]
    ENV[Environment Variables]
    YAML[YAML Config Files]
    Defaults[Default Values]
    
    CLI -->|Overrides| ENV
    ENV -->|Overrides| YAML
    YAML -->|Overrides| Defaults
    
    CLI -.->|Highest Priority| Final[Final Configuration]
    ENV -.-> Final
    YAML -.-> Final
    Defaults -.->|Lowest Priority| Final
```

## =� Configuration Structure

### Complete Configuration Schema
```yaml
# config.yaml - Complete configuration example
server:
  host: "0.0.0.0"
  port: "3002"
  read_timeout: "30s"
  write_timeout: "30s"
  shutdown_timeout: "10s"

security:
  enable_auth: true
  api_key: "${VIDEOCRAFT_SECURITY_API_KEY}"
  allowed_domains:
    - "trusted.example.com"
    - "api.trusted.org"
  enable_csrf: true
  csrf_secret: "${VIDEOCRAFT_SECURITY_CSRF_SECRET}"
  rate_limit: 100

python:
  path: "/usr/bin/python3"
  whisper_daemon_path: "./scripts/whisper_daemon.py"
  whisper_model: "base"
  whisper_device: "cpu"
  timeout: 300

ffmpeg:
  path: "/usr/bin/ffmpeg"
  timeout: 3600
  quality: "medium"
  preset: "medium"
  crf: 23

storage:
  output_dir: "./output"
  temp_dir: "./temp"
  max_age: 3600
  cleanup_interval: 300

subtitles:
  enabled: true
  style: "progressive"
  font_family: "Arial"
  font_size: 24
  position: "center-bottom"
  colors:
    word: "#FFFFFF"
    outline: "#000000"
    shadow: "#808080"

logging:
  level: "info"
  format: "json"
  output: "stdout"
  max_size: "100MB"
  max_backups: 5
  max_age: 30

monitoring:
  enable_metrics: true
  metrics_port: "9090"
  health_check_interval: "30s"
  enable_pprof: false
```

## < Environment Variables

### Naming Convention
Environment variables use the prefix `VIDEOCRAFT_` followed by the configuration path in uppercase with underscores:

```bash
# YAML: server.port -> ENV: VIDEOCRAFT_SERVER_PORT
# YAML: security.api_key -> ENV: VIDEOCRAFT_SECURITY_API_KEY
# YAML: python.whisper_model -> ENV: VIDEOCRAFT_PYTHON_WHISPER_MODEL
```

### Essential Environment Variables
```bash
# Server Configuration
export VIDEOCRAFT_SERVER_HOST="0.0.0.0"
export VIDEOCRAFT_SERVER_PORT="3002"

# Security Configuration (Required for production)
export VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
export VIDEOCRAFT_SECURITY_API_KEY="your-secure-api-key-here"
export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="trusted.com,api.trusted.com"
export VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
export VIDEOCRAFT_SECURITY_CSRF_SECRET="your-csrf-secret"

# Python/Whisper Configuration
export VIDEOCRAFT_PYTHON_PATH="/usr/bin/python3"
export VIDEOCRAFT_PYTHON_WHISPER_MODEL="base"
export VIDEOCRAFT_PYTHON_WHISPER_DEVICE="cpu"

# FFmpeg Configuration
export VIDEOCRAFT_FFMPEG_PATH="/usr/bin/ffmpeg"
export VIDEOCRAFT_FFMPEG_TIMEOUT=3600

# Storage Configuration
export VIDEOCRAFT_STORAGE_OUTPUT_DIR="./output"
export VIDEOCRAFT_STORAGE_TEMP_DIR="./temp"

# Logging Configuration
export VIDEOCRAFT_LOGGING_LEVEL="info"
export VIDEOCRAFT_LOGGING_FORMAT="json"
```

### Environment-Specific Configurations

#### Development Environment
```bash
# .env.development
VIDEOCRAFT_LOGGING_LEVEL=debug
VIDEOCRAFT_SECURITY_ENABLE_AUTH=false
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="localhost:3000,127.0.0.1:3000"
VIDEOCRAFT_SECURITY_ENABLE_CSRF=false
VIDEOCRAFT_MONITORING_ENABLE_PPROF=true
```

#### Staging Environment
```bash
# .env.staging
VIDEOCRAFT_LOGGING_LEVEL=info
VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
VIDEOCRAFT_SECURITY_API_KEY="staging-api-key"
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="staging.example.com"
VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
VIDEOCRAFT_MONITORING_ENABLE_METRICS=true
```

#### Production Environment
```bash
# .env.production
VIDEOCRAFT_LOGGING_LEVEL=warn
VIDEOCRAFT_SECURITY_ENABLE_AUTH=true
VIDEOCRAFT_SECURITY_API_KEY="${SECRET_MANAGER_API_KEY}"
VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="app.example.com"
VIDEOCRAFT_SECURITY_ENABLE_CSRF=true
VIDEOCRAFT_SECURITY_RATE_LIMIT=50
VIDEOCRAFT_MONITORING_ENABLE_METRICS=true
VIDEOCRAFT_MONITORING_ENABLE_PPROF=false
```

## =� Configuration Files

### File Discovery
VideoCraft searches for configuration files in the following locations:

1. `./config.yaml` (current directory)
2. `./config/config.yaml`
3. `/etc/videocraft/config.yaml`
4. `$HOME/.videocraft/config.yaml`

### Multiple Configuration Files
```bash
# Base configuration
config/
   base.yaml           # Common settings
   development.yaml    # Development overrides
   staging.yaml        # Staging overrides
   production.yaml     # Production overrides
```

```go
// Load configuration with environment-specific overrides
func LoadConfig(env string) (*Config, error) {
    viper.SetConfigName("base")
    viper.AddConfigPath("./config")
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    
    // Merge environment-specific config
    if env != "" {
        viper.SetConfigName(env)
        if err := viper.MergeInConfig(); err == nil {
            // Environment config merged successfully
        }
    }
    
    var config Config
    return &config, viper.Unmarshal(&config)
}
```

## = Security Configuration

### Authentication Settings
```yaml
security:
  enable_auth: true                    # Enable API authentication
  api_key: "${API_KEY}"               # API key (use env var)
  api_key_header: "Authorization"      # Header name for API key
  api_key_prefix: "Bearer"             # Prefix for bearer tokens
```

### CORS & CSRF Settings
```yaml
security:
  # CORS Configuration
  allowed_domains:                     # Explicit domain allowlist
    - "trusted.example.com"
    - "api.trusted.org"
  cors_max_age: 3600                   # Preflight cache duration
  cors_allow_credentials: true         # Allow credentials in CORS
  
  # CSRF Configuration
  enable_csrf: true                    # Enable CSRF protection
  csrf_secret: "${CSRF_SECRET}"        # CSRF secret (use env var)
  csrf_token_field: "csrf_token"       # Token field name in responses
  csrf_header_name: "X-CSRF-Token"     # Header name for CSRF token
```

### Rate Limiting
```yaml
security:
  rate_limit: 100                      # Requests per minute
  rate_limit_burst: 20                 # Burst allowance
  rate_limit_cleanup_interval: 60      # Cleanup interval (seconds)
```

## =� Storage Configuration

### File Storage Settings
```yaml
storage:
  output_dir: "./output"               # Final video output directory
  temp_dir: "./temp"                   # Temporary files directory
  max_age: 3600                        # File retention (seconds)
  cleanup_interval: 300                # Cleanup frequency (seconds)
  max_file_size: "500MB"               # Maximum upload size
  allowed_extensions:                  # Allowed file types
    - ".mp3"
    - ".wav"
    - ".png"
    - ".jpg"
    - ".mp4"
```

### Cloud Storage (Future)
```yaml
storage:
  provider: "local"                    # local, s3, gcs
  s3:
    bucket: "videocraft-storage"
    region: "us-west-2"
    access_key: "${AWS_ACCESS_KEY}"
    secret_key: "${AWS_SECRET_KEY}"
```

## <� Media Processing Configuration

### FFmpeg Settings
```yaml
ffmpeg:
  path: "/usr/bin/ffmpeg"              # FFmpeg binary path
  timeout: 3600                        # Processing timeout (seconds)
  
  # Video encoding settings
  quality: "medium"                    # low, medium, high
  preset: "medium"                     # FFmpeg preset
  crf: 23                              # Constant Rate Factor (18-28)
  resolution: "1920x1080"              # Default resolution
  framerate: 30                        # Output framerate
  
  # Audio settings
  audio_codec: "aac"
  audio_bitrate: "128k"
  audio_sample_rate: 44100
```

### Python/Whisper Settings
```yaml
python:
  path: "/usr/bin/python3"             # Python interpreter path
  whisper_daemon_path: "./scripts/whisper_daemon.py"
  whisper_model: "base"                # tiny, base, small, medium, large
  whisper_device: "cpu"                # cpu, cuda
  whisper_language: "auto"             # Language detection
  timeout: 300                         # Transcription timeout
  max_retries: 3                       # Retry attempts
```

### Subtitle Configuration
```yaml
subtitles:
  enabled: true
  style: "progressive"                 # progressive, classic
  
  # Font settings
  font_family: "Arial"
  font_size: 24
  position: "center-bottom"
  
  # Colors (hex format)
  colors:
    word: "#FFFFFF"                    # Text color
    outline: "#000000"                 # Outline color
    shadow: "#808080"                  # Shadow color
    background: "#000000"              # Background color
  
  # Styling
  outline_width: 2
  shadow_offset: 1
  line_spacing: 1.2
  margin: 10
```

## =� Logging Configuration

### Log Levels and Formats
```yaml
logging:
  level: "info"                        # debug, info, warn, error
  format: "json"                       # json, text
  output: "stdout"                     # stdout, stderr, file path
  
  # File logging (if output is file path)
  max_size: "100MB"                    # Log file size limit
  max_backups: 5                       # Number of backup files
  max_age: 30                          # Days to retain logs
  compress: true                       # Compress old logs
  
  # Structured logging fields
  include_caller: true                 # Include file:line info
  include_timestamp: true              # Include timestamps
  timestamp_format: "2006-01-02T15:04:05Z07:00"
```

### Component-Specific Logging
```yaml
logging:
  levels:
    api: "info"
    services: "info"
    ffmpeg: "warn"
    whisper: "error"
    security: "info"
```

## =� Monitoring Configuration

### Metrics and Health Checks
```yaml
monitoring:
  enable_metrics: true                 # Enable Prometheus metrics
  metrics_port: "9090"                 # Metrics endpoint port
  metrics_path: "/metrics"             # Metrics endpoint path
  
  # Health checks
  health_check_interval: "30s"         # Health check frequency
  health_timeout: "10s"                # Health check timeout
  
  # Performance profiling
  enable_pprof: false                  # Enable Go pprof (dev only)
  pprof_port: "6060"                   # pprof endpoint port
```

## =� Runtime Configuration

### Command Line Interface
```bash
# Start with custom config file
./videocraft --config=/path/to/config.yaml

# Override specific settings
./videocraft --server.port=8080 --logging.level=debug

# Environment-specific configuration
./videocraft --env=production

# Validate configuration without starting
./videocraft --validate-config
```

### Configuration Validation
```bash
# Validate current configuration
curl http://localhost:3002/api/v1/config/validate

# Get current configuration (sanitized)
curl http://localhost:3002/api/v1/config
```

## =� Configuration Tools

### Environment Variable Generator
```bash
#!/bin/bash
# generate-env.sh - Generate environment variables from YAML

yq eval '.security.api_key' config.yaml | \
  sed 's/^/export VIDEOCRAFT_SECURITY_API_KEY=/' > .env

yq eval '.security.allowed_domains[]' config.yaml | \
  tr '\n' ',' | sed 's/,$//' | \
  sed 's/^/export VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS="/' | \
  sed 's/$/"/' >> .env
```

### Configuration Validation Script
```bash
#!/bin/bash
# validate-config.sh - Validate configuration before deployment

echo "Validating VideoCraft configuration..."

# Check required environment variables
required_vars=(
  "VIDEOCRAFT_SECURITY_API_KEY"
  "VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS"
)

for var in "${required_vars[@]}"; do
  if [[ -z "${!var}" ]]; then
    echo "ERROR: Required environment variable $var is not set"
    exit 1
  fi
done

# Test configuration loading
if ./videocraft --validate-config; then
  echo "Configuration validation passed"
else
  echo "Configuration validation failed"
  exit 1
fi
```

## = Troubleshooting Configuration

### Common Issues

#### Configuration Not Loading
**Problem**: Settings not applied
**Solution**: Check file paths and environment variable names
```bash
# Debug configuration loading
export VIDEOCRAFT_LOGGING_LEVEL=debug
./videocraft 2>&1 | grep -i config
```

#### Environment Variable Override
**Problem**: YAML values not overridden by environment
**Solution**: Verify environment variable naming
```bash
# Check environment variables
env | grep VIDEOCRAFT_

# Test specific variable
echo $VIDEOCRAFT_SECURITY_API_KEY
```

#### Path Resolution Issues
**Problem**: File paths not found
**Solution**: Use absolute paths or verify working directory
```yaml
# Use absolute paths in production
ffmpeg:
  path: "/usr/local/bin/ffmpeg"  # Absolute path
python:
  path: "/opt/python3/bin/python3"  # Absolute path
```

### Debug Commands
```bash
# Show effective configuration
curl -H "Authorization: Bearer $API_KEY" \
  http://localhost:3002/api/v1/config

# Test specific components
curl http://localhost:3002/health/detailed

# Validate configuration syntax
yq validate config.yaml
```

## =� Related Topics

### Security Configuration
- **[Security Overview](../security/overview.md)** - Security architecture
- **[Authentication](../security/authentication.md)** - API key configuration
- **[CORS & CSRF](../security/cors-csrf.md)** - Cross-origin security

### Deployment Configuration
- **[Docker Setup](../deployment/docker.md)** - Container configuration
- **[Kubernetes Setup](../deployment/kubernetes.md)** - K8s configuration
- **[Production Setup](../deployment/production-setup.md)** - Production configuration

### Environment-Specific Guides
- **[Environment Variables](environment-variables.md)** - Complete variable reference
- **[Service Configuration](service-configuration.md)** - Service-specific settings
- **[Performance Tuning](performance-tuning.md)** - Performance optimization

---

**= Next Steps**: [Environment Variables](environment-variables.md) | [Security Configuration](../security/overview.md) | [Production Setup](../deployment/production-setup.md)