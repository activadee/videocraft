package transcription

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/activadee/videocraft/internal/app"
	"github.com/activadee/videocraft/internal/pkg/errors"
	"github.com/activadee/videocraft/internal/pkg/logger"
)

// Service provides transcription capabilities using Whisper AI
type Service interface {
	TranscribeAudio(ctx context.Context, audioURL string) (*TranscriptionResult, error)
	StartDaemon() error
	StopDaemon() error
	HealthCheck() error
	Shutdown()
}

type service struct {
	cfg    *app.Config
	log    logger.Logger
	daemon *WhisperDaemon
	mutex  sync.RWMutex
}

// NewService creates a new transcription service
func NewService(cfg *app.Config, log logger.Logger) Service {
	return &service{
		cfg:    cfg,
		log:    log,
		daemon: nil,
		mutex:  sync.RWMutex{},
	}
}

type WhisperDaemon struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	scanner *bufio.Scanner

	cfg          *app.Config
	log          logger.Logger
	running      bool
	mutex        sync.RWMutex
	restartCount int
	lastRestart  time.Time
}

type TranscriptionRequest struct {
	ID             string `json:"id"`
	Action         string `json:"action"`
	URL            string `json:"url,omitempty"`
	Language       string `json:"language,omitempty"`
	WordTimestamps bool   `json:"word_timestamps,omitempty"`
}

type TranscriptionResponse struct {
	ID             string                 `json:"id"`
	Success        bool                   `json:"success"`
	Text           string                 `json:"text,omitempty"`
	Language       string                 `json:"language,omitempty"`
	Duration       float64                `json:"duration,omitempty"`
	Segments       []WhisperSegment       `json:"segments,omitempty"`
	WordTimestamps []WhisperWordTimestamp `json:"word_timestamps,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Traceback      string                 `json:"traceback,omitempty"`

	// Status response fields
	ModelLoaded bool   `json:"model_loaded,omitempty"`
	Model       string `json:"model,omitempty"`
	Device      string `json:"device,omitempty"`
	Message     string `json:"message,omitempty"`
}

type WhisperSegment struct {
	Start float64                `json:"start"`
	End   float64                `json:"end"`
	Text  string                 `json:"text"`
	Words []WhisperWordTimestamp `json:"words,omitempty"`
}

type WhisperWordTimestamp struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// TranscriptionResult represents the result of audio transcription
type TranscriptionResult struct {
	Text           string                 `json:"text"`
	Language       string                 `json:"language"`
	Duration       float64                `json:"duration"`
	WordTimestamps []WhisperWordTimestamp `json:"word_timestamps"`
	Success        bool                   `json:"success"`
}

// Deprecated: Use NewService instead
func newTranscriptionService(cfg *app.Config, log logger.Logger) Service {
	return NewService(cfg, log)
}

func (ts *service) TranscribeAudio(ctx context.Context, url string) (*TranscriptionResult, error) {
	ts.log.Debugf("Transcribing audio: %s", url)

	if !ts.cfg.Transcription.Enabled {
		ts.log.Debug("Transcription disabled in configuration")
		return nil, errors.InvalidInput("transcription is disabled")
	}

	// Only use daemon mode - no fallback
	if !ts.cfg.Transcription.Daemon.Enabled {
		return nil, errors.InvalidInput("daemon mode is required but disabled")
	}

	return ts.transcribeWithDaemon(ctx, url)
}

func (ts *service) transcribeWithDaemon(ctx context.Context, url string) (*TranscriptionResult, error) {
	// Ensure daemon is running
	if err := ts.ensureDaemon(); err != nil {
		return nil, fmt.Errorf("failed to start daemon: %w", err)
	}

	// Create request
	request := TranscriptionRequest{
		ID:             uuid.New().String(),
		Action:         "transcribe",
		URL:            url,
		Language:       ts.cfg.Transcription.Python.Language,
		WordTimestamps: true,
	}

	// Send request to daemon
	ts.daemon.mutex.Lock()
	defer ts.daemon.mutex.Unlock()

	if !ts.daemon.running {
		return nil, fmt.Errorf("daemon not running")
	}

	// Send JSON request
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	_, err = ts.daemon.stdin.Write(append(requestJSON, '\n'))
	if err != nil {
		return nil, fmt.Errorf("failed to send request to daemon: %w", err)
	}

	// Read response with timeout
	responseCtx, cancel := context.WithTimeout(ctx, ts.cfg.Transcription.Processing.Timeout)
	defer cancel()

	responseChan := make(chan TranscriptionResponse, 1)
	errorChan := make(chan error, 1)

	go func() {
		if !ts.daemon.scanner.Scan() {
			if err := ts.daemon.scanner.Err(); err != nil {
				errorChan <- err
			} else {
				errorChan <- fmt.Errorf("unexpected EOF from daemon")
			}
			return
		}

		var response TranscriptionResponse
		if err := json.Unmarshal([]byte(ts.daemon.scanner.Text()), &response); err != nil {
			errorChan <- fmt.Errorf("failed to parse response: %w", err)
			return
		}

		responseChan <- response
	}()

	select {
	case response := <-responseChan:
		if !response.Success {
			return nil, fmt.Errorf("transcription failed: %s", response.Error)
		}
		return ts.convertToTranscriptionResult(response), nil
	case err := <-errorChan:
		return nil, err
	case <-responseCtx.Done():
		return nil, fmt.Errorf("transcription timeout")
	}
}

func (ts *service) ensureDaemon() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	// Check if daemon is already running
	if ts.daemon != nil && ts.daemon.running {
		return nil
	}

	// Start new daemon
	return ts.startDaemon()
}

func (ts *service) startDaemon() error {
	ts.log.Info("Starting Whisper daemon")

	// Build command
	scriptPath := filepath.Join(ts.cfg.Transcription.Python.ScriptPath, "whisper_daemon.py")
	idleTimeout := int(ts.cfg.Transcription.Daemon.IdleTimeout.Seconds())

	cmd := exec.Command(ts.cfg.Transcription.Python.Path, scriptPath,
		"--idle-timeout", fmt.Sprintf("%d", idleTimeout),
		"--model", ts.cfg.Transcription.Python.Model,
		"--log-level", "INFO",
	)

	// Setup pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Create daemon object
	daemon := &WhisperDaemon{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		scanner: bufio.NewScanner(stdout),
		cfg:     ts.cfg,
		log:     ts.log,
		running: true,
	}

	ts.daemon = daemon

	// Start monitoring goroutines
	go ts.monitorDaemon()
	go ts.logDaemonErrors()

	// Wait for daemon to be ready (with timeout)
	if err := ts.waitForDaemonReady(); err != nil {
		ts.stopDaemon()
		return fmt.Errorf("daemon startup failed: %w", err)
	}

	ts.log.Info("Whisper daemon started successfully")
	return nil
}

func (ts *service) waitForDaemonReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), ts.cfg.Transcription.Daemon.StartupTimeout)
	defer cancel()

	ts.log.Info("Waiting for daemon to load Whisper model...")

	// Send status request to ensure model is loaded
	request := TranscriptionRequest{
		ID:     uuid.New().String(),
		Action: "status",
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal status request: %w", err)
	}

	_, err = ts.daemon.stdin.Write(append(requestJSON, '\n'))
	if err != nil {
		return fmt.Errorf("failed to send status: %w", err)
	}

	// Wait for response with model loaded confirmation
	responseChan := make(chan bool, 1)
	errorChan := make(chan error, 1)

	go func() {
		for {
			if !ts.daemon.scanner.Scan() {
				if err := ts.daemon.scanner.Err(); err != nil {
					errorChan <- err
				} else {
					errorChan <- fmt.Errorf("daemon closed unexpectedly")
				}
				return
			}

			responseText := ts.daemon.scanner.Text()
			ts.log.Debugf("Daemon response: %s", responseText)

			var response TranscriptionResponse
			if err := json.Unmarshal([]byte(responseText), &response); err != nil {
				ts.log.Debugf("Failed to parse response as JSON: %v", err)
				continue // Skip non-JSON output (like warnings)
			}

			if response.Success && response.ModelLoaded {
				// Model is loaded and daemon is ready
				responseChan <- true
				return
			} else if response.Success && !response.ModelLoaded {
				// Daemon responded but model not loaded yet, keep waiting
				ts.log.Debug("Daemon running but model not loaded yet, waiting...")
				continue
			}
		}
	}()

	select {
	case <-responseChan:
		ts.log.Info("Daemon is ready with model loaded")
		return nil
	case err := <-errorChan:
		return fmt.Errorf("daemon error during startup: %w", err)
	case <-ctx.Done():
		return fmt.Errorf("daemon startup timeout - model loading took too long")
	}
}

func (ts *service) monitorDaemon() {
	if ts.daemon == nil {
		return
	}

	// Wait for process to exit
	err := ts.daemon.cmd.Wait()

	ts.daemon.mutex.Lock()
	ts.daemon.running = false
	ts.daemon.mutex.Unlock()

	if err != nil {
		ts.log.Errorf("Whisper daemon exited with error: %v", err)
	} else {
		ts.log.Info("Whisper daemon exited normally")
	}

	// Attempt restart if within limits
	if ts.shouldRestartDaemon() {
		ts.log.Info("Attempting to restart Whisper daemon")
		time.Sleep(time.Second * 5) // Brief delay before restart
		if err := ts.ensureDaemon(); err != nil {
			ts.log.Errorf("Failed to restart daemon: %v", err)
		}
	}
}

func (ts *service) logDaemonErrors() {
	if ts.daemon == nil || ts.daemon.stderr == nil {
		return
	}

	scanner := bufio.NewScanner(ts.daemon.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			ts.log.Debugf("Daemon stderr: %s", line)
		}
	}
}

func (ts *service) shouldRestartDaemon() bool {
	now := time.Now()
	if now.Sub(ts.daemon.lastRestart) > time.Minute*5 {
		ts.daemon.restartCount = 0 // Reset counter after 5 minutes
	}

	ts.daemon.restartCount++
	ts.daemon.lastRestart = now

	return ts.daemon.restartCount <= ts.cfg.Transcription.Daemon.RestartMaxAttempts
}

func (ts *service) stopDaemon() {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.daemon == nil {
		return
	}

	ts.log.Info("Stopping Whisper daemon")

	// Send shutdown command
	if ts.daemon.running {
		shutdownRequest := TranscriptionRequest{
			ID:     uuid.New().String(),
			Action: "shutdown",
		}
		if requestJSON, err := json.Marshal(shutdownRequest); err == nil {
			if _, writeErr := ts.daemon.stdin.Write(append(requestJSON, '\n')); writeErr != nil {
				ts.log.Errorf("Failed to write shutdown request: %v", writeErr)
			}
		}
	}

	// Close pipes
	if ts.daemon.stdin != nil {
		ts.daemon.stdin.Close()
	}
	if ts.daemon.stdout != nil {
		ts.daemon.stdout.Close()
	}
	if ts.daemon.stderr != nil {
		ts.daemon.stderr.Close()
	}

	// Wait for process to exit (with timeout)
	done := make(chan error, 1)
	go func() {
		done <- ts.daemon.cmd.Wait()
	}()

	select {
	case <-done:
		ts.log.Info("Daemon stopped gracefully")
	case <-time.After(10 * time.Second):
		ts.log.Warn("Daemon shutdown timeout, killing process")
		if killErr := ts.daemon.cmd.Process.Kill(); killErr != nil {
			ts.log.Errorf("Failed to kill daemon process: %v", killErr)
		}
	}

	ts.daemon = nil
}

func (ts *service) convertToTranscriptionResult(response TranscriptionResponse) *TranscriptionResult {
	return &TranscriptionResult{
		Text:           response.Text,
		Language:       response.Language,
		Duration:       response.Duration,
		WordTimestamps: response.WordTimestamps,
		Success:        response.Success,
	}
}

func (ts *service) Shutdown() {
	ts.stopDaemon()
}

func (ts *service) HealthCheck() error {
	if ts.daemon == nil {
		return fmt.Errorf("transcription daemon not initialized")
	}

	ts.daemon.mutex.RLock()
	running := ts.daemon.running
	ts.daemon.mutex.RUnlock()

	if !running {
		return fmt.Errorf("transcription daemon not running")
	}

	return nil
}

func (ts *service) StartDaemon() error {
	return ts.ensureDaemon()
}

func (ts *service) StopDaemon() error {
	ts.stopDaemon()
	return nil
}
