package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

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

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type FFmpegConfig struct {
	BinaryPath  string        `mapstructure:"binary_path"`
	FFprobePath string        `mapstructure:"ffprobe_path"`
	Timeout     time.Duration `mapstructure:"timeout"`
	Quality     int           `mapstructure:"quality"`
	Preset      string        `mapstructure:"preset"`
}

type TranscriptionConfig struct {
	Enabled    bool             `mapstructure:"enabled"`
	Daemon     DaemonConfig     `mapstructure:"daemon"`
	Python     PythonConfig     `mapstructure:"python"`
	Processing ProcessingConfig `mapstructure:"processing"`
}

type DaemonConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
	StartupTimeout     time.Duration `mapstructure:"startup_timeout"`
	RestartMaxAttempts int           `mapstructure:"restart_max_attempts"`
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

type StorageConfig struct {
	OutputDir       string        `mapstructure:"output_dir"`
	TempDir         string        `mapstructure:"temp_dir"`
	MaxFileSize     int64         `mapstructure:"max_file_size"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	RetentionDays   int           `mapstructure:"retention_days"`
}

type JobConfig struct {
	Workers             int           `mapstructure:"workers"`
	QueueSize           int           `mapstructure:"queue_size"`
	MaxConcurrent       int           `mapstructure:"max_concurrent"`
	StatusCheckInterval time.Duration `mapstructure:"status_check_interval"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type SecurityConfig struct {
	APIKey         string   `mapstructure:"api_key"`
	RateLimit      int      `mapstructure:"rate_limit"`
	EnableAuth     bool     `mapstructure:"enable_auth"`
	AllowedDomains []string `mapstructure:"allowed_domains"`
	EnableCSRF     bool     `mapstructure:"enable_csrf"`
	CSRFSecret     string   `mapstructure:"csrf_secret"`
}

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
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Log all VIDEOCRAFT environment variables for debugging
	logEnvironmentVariables()

	// Manual bindings for complex types
	_ = viper.BindEnv("security.allowed_domains", "VIDEOCRAFT_SECURITY_ALLOWED_DOMAINS")

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

	// Generate API key if authentication is enabled but no key is provided
	// Only generate if no API key is configured via environment or config file
	if config.Security.EnableAuth && config.Security.APIKey == "" && !viper.IsSet("security.api_key") {
		generatedKey, err := generateSecureAPIKey()
		// If generation fails, return an error
		if err != nil {
			return nil, fmt.Errorf("failed to generate API key: %w", err)
		}
		fmt.Printf("Generated API key: %s\n", generatedKey) // Log the generated key for visibility
		config.Security.APIKey = generatedKey
	}

	// Auto-generate strong CSRF secret if CSRF is enabled but none supplied
	if config.Security.EnableCSRF &&
		config.Security.CSRFSecret == "" &&
		!viper.IsSet("security.csrf_secret") {
		secret, err := generateSecureAPIKey() // 256-bit hex == 64 chars
		if err != nil {
			return nil, fmt.Errorf("failed to generate CSRF secret: %w", err)
		}
		config.Security.CSRFSecret = secret
	}

	return &config, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3002)

	// FFmpeg defaults
	viper.SetDefault("ffmpeg.binary_path", "ffmpeg")
	viper.SetDefault("ffmpeg.ffprobe_path", "ffprobe")
	viper.SetDefault("ffmpeg.timeout", "1h")
	viper.SetDefault("ffmpeg.quality", 23)
	viper.SetDefault("ffmpeg.preset", "medium")

	// Transcription defaults
	viper.SetDefault("transcription.enabled", true)
	viper.SetDefault("transcription.daemon.enabled", true)
	viper.SetDefault("transcription.daemon.idle_timeout", "300s")
	viper.SetDefault("transcription.daemon.startup_timeout", "30s")
	viper.SetDefault("transcription.daemon.restart_max_attempts", 3)
	viper.SetDefault("transcription.python.path", "python3")
	viper.SetDefault("transcription.python.script_path", "./scripts")
	viper.SetDefault("transcription.python.model", "base")
	viper.SetDefault("transcription.python.language", "auto")
	viper.SetDefault("transcription.python.device", "auto")
	viper.SetDefault("transcription.processing.workers", 2)
	viper.SetDefault("transcription.processing.timeout", "60s")

	// Subtitles defaults
	viper.SetDefault("subtitles.enabled", true)
	viper.SetDefault("subtitles.style", "progressive")
	viper.SetDefault("subtitles.font_family", "Arial")
	viper.SetDefault("subtitles.font_size", 24)
	viper.SetDefault("subtitles.position", "center-bottom")
	viper.SetDefault("subtitles.colors.word", "#FFFFFF")
	viper.SetDefault("subtitles.colors.outline", "#000000")

	// Storage defaults
	viper.SetDefault("storage.output_dir", "./generated_videos")
	viper.SetDefault("storage.temp_dir", "./temp")
	viper.SetDefault("storage.max_file_size", 1073741824) // 1GB
	viper.SetDefault("storage.cleanup_interval", "1h")
	viper.SetDefault("storage.retention_days", 7)

	// Job defaults
	viper.SetDefault("job.workers", 4)
	viper.SetDefault("job.queue_size", 100)
	viper.SetDefault("job.max_concurrent", 10)
	viper.SetDefault("job.status_check_interval", "5s")

	// Log defaults
	viper.SetDefault("log.level", "debug")
	viper.SetDefault("log.format", "text")

	// Security defaults - CRITICAL: Authentication enabled by default for security
	viper.SetDefault("security.rate_limit", 100)
	viper.SetDefault("security.enable_auth", true)
	viper.SetDefault("security.allowed_domains", []string{})
	viper.SetDefault("security.enable_csrf", false)
	viper.SetDefault("security.csrf_secret", "CHANGE_ME_64_CHAR_MINIMUM_ENTROPY_SECRET_FOR_CSRF_PROTECTION_REPLACE")
}

// generateSecureAPIKey generates a cryptographically secure API key
func generateSecureAPIKey() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Encode as hexadecimal string (64 characters)
	return hex.EncodeToString(bytes), nil
}

// logEnvironmentVariables logs all VIDEOCRAFT environment variables for debugging
func logEnvironmentVariables() {
	fmt.Println("=== VIDEOCRAFT Environment Variables Debug ===")
	
	var videocraftVars []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "VIDEOCRAFT_") {
			videocraftVars = append(videocraftVars, env)
		}
	}
	
	if len(videocraftVars) == 0 {
		fmt.Println("No VIDEOCRAFT_ environment variables found")
		return
	}
	
	sort.Strings(videocraftVars)
	for _, env := range videocraftVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			
			// Mask sensitive values
			if strings.Contains(strings.ToLower(key), "key") || 
			   strings.Contains(strings.ToLower(key), "secret") ||
			   strings.Contains(strings.ToLower(key), "password") {
				if len(value) > 8 {
					value = value[:4] + "***" + value[len(value)-4:]
				} else {
					value = "***"
				}
			}
			
			fmt.Printf("  %s = %s\n", key, value)
		}
	}
	fmt.Println("=== End Environment Variables ===")
}
