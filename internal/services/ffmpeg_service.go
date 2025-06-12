package services

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/activadee/videocraft/internal/config"
	"github.com/activadee/videocraft/internal/domain/errors"
	"github.com/activadee/videocraft/internal/domain/models"
	"github.com/activadee/videocraft/pkg/logger"
)

type ffmpegService struct {
	cfg           *config.Config
	log           logger.Logger
	transcription TranscriptionService
	subtitle      SubtitleService
	audio         AudioService
}

func (s *ffmpegService) GenerateVideo(ctx context.Context, config *models.VideoConfigArray, progressChan chan<- int) (string, error) {
	s.log.Info("Starting video generation")

	// Build FFmpeg command with optional subtitles
	cmd, err := s.BuildCommandWithSubtitles(ctx, config)
	if err != nil {
		return "", errors.FFmpegFailed(fmt.Errorf("failed to build command: %w", err))
	}

	s.log.Debugf("Generated FFmpeg command: %s %s", s.cfg.FFmpeg.BinaryPath, strings.Join(cmd.Args, " "))

	// Execute command with timeout
	ctx, cancel := context.WithTimeout(ctx, s.cfg.FFmpeg.Timeout)
	defer cancel()

	ffmpegCmd := exec.CommandContext(ctx, s.cfg.FFmpeg.BinaryPath, cmd.Args...)

	// Setup progress tracking
	if progressChan != nil {
		stderr, err := ffmpegCmd.StderrPipe()
		if err != nil {
			return "", errors.FFmpegFailed(err)
		}

		// Parse progress in goroutine
		go s.parseProgress(stderr, progressChan)
	}

	// Execute command
	if err := ffmpegCmd.Run(); err != nil {
		return "", errors.FFmpegFailed(err)
	}

	s.log.Infof("Video generation completed: %s", cmd.OutputPath)
	return cmd.OutputPath, nil
}

func (s *ffmpegService) BuildCommandWithSubtitles(ctx context.Context, config *models.VideoConfigArray) (*FFmpegCommand, error) {
	if len(*config) == 0 {
		return nil, fmt.Errorf("no video projects provided")
	}

	project := (*config)[0]
	
	// Generate subtitles if enabled and subtitle element exists
	var subtitleFilePath string
	var actualTotalDuration float64
	
	if s.cfg.Subtitles.Enabled && s.hasSubtitleElement(project) {
		s.log.Info("Generating subtitles for video")
		
		subtitleResult, err := s.subtitle.GenerateSubtitles(ctx, project)
		if err != nil {
			s.log.Warnf("Failed to generate subtitles: %v", err)
			// Continue without subtitles - use fallback duration calculation
			actualTotalDuration = s.calculateFallbackDuration(project)
		} else if subtitleResult != nil {
			subtitleFilePath = subtitleResult.FilePath
			actualTotalDuration = subtitleResult.TotalDuration.Seconds()
			s.log.Infof("Subtitles generated: %s (%d events, %s style), total duration: %.2fs", 
				subtitleFilePath, subtitleResult.EventCount, subtitleResult.Style, actualTotalDuration)
		}
	} else {
		// No subtitles - use fallback duration calculation
		actualTotalDuration = s.calculateFallbackDuration(project)
	}
	
	return s.buildCommandWithSubtitleFileAndDuration(config, subtitleFilePath, actualTotalDuration)
}

func (s *ffmpegService) BuildCommand(config *models.VideoConfigArray) (*FFmpegCommand, error) {
	if len(*config) == 0 {
		return nil, fmt.Errorf("no video projects provided")
	}

	// Security validation: Check all URLs in configuration
	if err := s.validateAllURLsInConfig(config); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// For now, process the first project in the array
	project := (*config)[0]
	
	builder := newCommandBuilder()

	// Find background video element
	var backgroundVideo *models.Element
	for _, element := range project.Elements {
		if element.Type == "video" {
			backgroundVideo = &element
			break
		}
	}

	if backgroundVideo == nil {
		return nil, fmt.Errorf("no background video element found")
	}

	// Collect all audio elements from scenes
	audioElements := s.collectAudioElements(project)
	
	// Collect all image elements from scenes
	imageElements := s.collectImageElements(project)

	// Calculate total duration
	totalDuration := s.calculateTotalDuration(audioElements)

	// Add inputs
	builder.addInput("-y") // Overwrite output
	builder.addInput("-protocol_whitelist", "file,http,https,tcp,tls")
	
	// Background video with loop
	loopsNeeded := int(totalDuration/backgroundVideo.Duration) + 1
	builder.addInput("-stream_loop", fmt.Sprintf("%d", loopsNeeded), "-i", backgroundVideo.Src)

	// Audio inputs
	for _, audio := range audioElements {
		builder.addInput("-i", audio.Src)
	}

	// Image inputs
	for _, image := range imageElements {
		builder.addInput("-i", image.Src)
	}

	// Build filter complex with proper scene timing
	sceneTiming, err := s.analyzeSceneTiming(audioElements)
	if err != nil {
		s.log.Warnf("Failed to analyze scene timing: %v, using fallback", err)
		sceneTiming = s.generateFallbackTiming(audioElements)
	}
	filterComplex := s.buildFilterComplexWithSceneTiming(project, audioElements, imageElements, sceneTiming, totalDuration)

	if filterComplex != "" {
		builder.addArg("-filter_complex", filterComplex)
	}

	// Map outputs
	if len(imageElements) > 0 {
		builder.addArg("-map", fmt.Sprintf("[overlay_%d]", len(imageElements)-1))
	} else {
		builder.addArg("-map", "0:v")
	}

	if len(audioElements) > 0 {
		builder.addArg("-map", "[final_audio]")
	}

	// Set duration
	builder.addArg("-t", fmt.Sprintf("%.2f", totalDuration))

	// Output settings based on project config
	s.addOutputSettingsForProject(builder, project)

	// Generate output path
	outputPath := s.generateOutputPathForProject(project)
	builder.addArg(outputPath)

	return &FFmpegCommand{
		Args:       builder.args,
		OutputPath: outputPath,
	}, nil
}

func (s *ffmpegService) Execute(ctx context.Context, cmd *FFmpegCommand) error {
	ffmpegCmd := exec.CommandContext(ctx, s.cfg.FFmpeg.BinaryPath, cmd.Args...)
	return ffmpegCmd.Run()
}


func (s *ffmpegService) parseProgress(stderr io.ReadCloser, progressChan chan<- int) {
	defer close(progressChan)
	defer stderr.Close()

	scanner := bufio.NewScanner(stderr)
	var totalDuration float64
	
	// Regular expressions for parsing FFmpeg output
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	timeRegex := regexp.MustCompile(`time=(\d{2}):(\d{2}):(\d{2})\.(\d{2})`)
	
	for scanner.Scan() {
		line := scanner.Text()
		s.log.Debugf("FFmpeg output: %s", line)
		
		// Parse total duration from the beginning
		if totalDuration == 0 {
			if matches := durationRegex.FindStringSubmatch(line); len(matches) == 5 {
				hours, _ := strconv.Atoi(matches[1])
				minutes, _ := strconv.Atoi(matches[2])
				seconds, _ := strconv.Atoi(matches[3])
				centiseconds, _ := strconv.Atoi(matches[4])
				
				totalDuration = float64(hours*3600 + minutes*60 + seconds) + float64(centiseconds)/100
				s.log.Debugf("Total duration parsed: %.2f seconds", totalDuration)
			}
		}
		
		// Parse current time progress
		if totalDuration > 0 {
			if matches := timeRegex.FindStringSubmatch(line); len(matches) == 5 {
				hours, _ := strconv.Atoi(matches[1])
				minutes, _ := strconv.Atoi(matches[2])
				seconds, _ := strconv.Atoi(matches[3])
				centiseconds, _ := strconv.Atoi(matches[4])
				
				currentTime := float64(hours*3600 + minutes*60 + seconds) + float64(centiseconds)/100
				progress := int((currentTime / totalDuration) * 100)
				
				// Cap progress at 100%
				if progress > 100 {
					progress = 100
				}
				
				// Send progress update
				select {
				case progressChan <- progress:
					s.log.Debugf("Progress update: %d%%", progress)
				default:
				}
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		s.log.Errorf("Error reading FFmpeg stderr: %v", err)
	}
}

// Command builder helper
type commandBuilder struct {
	args []string
}

func newCommandBuilder() *commandBuilder {
	return &commandBuilder{
		args: []string{"-y"}, // Overwrite output
	}
}

func (cb *commandBuilder) addInput(args ...string) {
	cb.args = append(cb.args, args...)
}

func (cb *commandBuilder) addArg(args ...string) {
	cb.args = append(cb.args, args...)
}


// Helper functions for new scene-based architecture

func (s *ffmpegService) collectAudioElements(project models.VideoProject) []models.Element {
	var audioElements []models.Element
	
	// Collect from scenes in order
	for _, scene := range project.Scenes {
		for _, element := range scene.Elements {
			if element.Type == "audio" {
				audioElements = append(audioElements, element)
			}
		}
	}
	
	return audioElements
}

func (s *ffmpegService) collectImageElements(project models.VideoProject) []models.Element {
	var imageElements []models.Element
	
	// Collect from scenes in order
	for _, scene := range project.Scenes {
		for _, element := range scene.Elements {
			if element.Type == "image" {
				imageElements = append(imageElements, element)
			}
		}
	}
	
	return imageElements
}

func (s *ffmpegService) calculateTotalDuration(audioElements []models.Element) float64 {
	var total float64
	for _, audio := range audioElements {
		if audio.Duration > 0 {
			total += audio.Duration
		}
	}
	// Add 2 second buffer like in Python implementation
	return total + 2.0
}

func (s *ffmpegService) calculateFallbackDuration(project models.VideoProject) float64 {
	// Fallback: Use background video duration if available
	for _, element := range project.Elements {
		if element.Type == "video" && element.Duration > 0 {
			s.log.Warnf("Using fallback duration from background video: %.2fs", element.Duration)
			return element.Duration
		}
	}
	
	// Last resort: default duration
	s.log.Warn("No duration information available, using default 30 seconds")
	return 30.0
}

func (s *ffmpegService) buildFilterComplexWithSceneTiming(project models.VideoProject, audioElements, imageElements []models.Element, sceneTiming []models.TimingSegment, totalDuration float64) string {
	var filters []string
	
	// Audio concatenation
	if len(audioElements) > 1 {
		audioInputs := make([]string, len(audioElements))
		for i := range audioElements {
			audioInputs[i] = fmt.Sprintf("[%d:a]", i+1) // +1 because 0 is background video
		}
		audioConcat := fmt.Sprintf("%sconcat=n=%d:v=0:a=1[concatenated_audio]",
			strings.Join(audioInputs, ""),
			len(audioElements))
		filters = append(filters, audioConcat)
		filters = append(filters, "[concatenated_audio]apad=pad_dur=2[final_audio]")
	} else if len(audioElements) == 1 {
		filters = append(filters, "[1:a]apad=pad_dur=2[final_audio]")
	}
	
	// Image overlays with timing based on actual audio analysis
	currentInput := "0:v"
	
	for i, image := range imageElements {
		// Use scene timing from audio analysis
		var startTime, endTime float64
		if i < len(sceneTiming) {
			startTime = sceneTiming[i].StartTime
			endTime = sceneTiming[i].EndTime
		} else {
			// Fallback if we have more images than timing segments
			startTime = float64(i) * 5.0
			endTime = startTime + 5.0
		}
		
		s.log.Debugf("Image %d overlay timing: %.2fs - %.2fs (duration: %.2fs)", 
			i, startTime, endTime, endTime-startTime)
		
		// Scale image - use correct input index for images with :v selector
		imageInputIndex := len(audioElements) + 1 + i
		scaleFilter := fmt.Sprintf("[%d:v]scale=500:500[scaled_img_%d]",
			imageInputIndex, i)
		filters = append(filters, scaleFilter)
		
		// Overlay with timing based on actual audio duration
		overlayFilter := fmt.Sprintf("[%s][scaled_img_%d]overlay=%d:%d:enable='between(t\\,%f\\,%f)'[overlay_%d]",
			currentInput, i, image.X, image.Y, startTime, endTime, i)
		filters = append(filters, overlayFilter)
		
		currentInput = fmt.Sprintf("overlay_%d", i)
	}
	
	return strings.Join(filters, ";")
}

func (s *ffmpegService) addOutputSettingsForProject(builder *commandBuilder, project models.VideoProject) {
	// Codec settings
	builder.addArg("-c:v", "libx264")
	builder.addArg("-c:a", "aac")
	
	// Quality based on project settings
	if project.Quality == "high" {
		builder.addArg("-crf", "18")
	} else {
		builder.addArg("-crf", "23")
	}
	
	// Resolution
	if project.Width > 0 && project.Height > 0 {
		builder.addArg("-s", fmt.Sprintf("%dx%d", project.Width, project.Height))
	}
	
	// Additional settings
	builder.addArg("-preset", "medium")
	builder.addArg("-movflags", "+faststart")
	builder.addArg("-pix_fmt", "yuv420p")
}

func (s *ffmpegService) generateOutputPathForProject(project models.VideoProject) string {
	format := "mp4" // default format
	filename := fmt.Sprintf("video_%s.%s", uuid.New().String()[:8], format)
	return filepath.Join(s.cfg.Storage.OutputDir, filename)
}

func (s *ffmpegService) hasSubtitleElement(project models.VideoProject) bool {
	for _, element := range project.Elements {
		if element.Type == "subtitles" {
			return true
		}
	}
	return false
}

func (s *ffmpegService) buildCommandWithSubtitleFileAndDuration(config *models.VideoConfigArray, subtitleFilePath string, totalDuration float64) (*FFmpegCommand, error) {
	if len(*config) == 0 {
		return nil, fmt.Errorf("no video projects provided")
	}

	// For now, process the first project in the array
	project := (*config)[0]
	
	builder := newCommandBuilder()

	// Find background video element
	var backgroundVideo *models.Element
	for _, element := range project.Elements {
		if element.Type == "video" {
			backgroundVideo = &element
			break
		}
	}

	if backgroundVideo == nil {
		return nil, fmt.Errorf("no background video element found")
	}

	// Collect all audio elements from scenes
	audioElements := s.collectAudioElements(project)
	
	// Collect all image elements from scenes
	imageElements := s.collectImageElements(project)

	// Analyze audio timing for scene-based overlays using AudioService
	sceneTiming, err := s.analyzeSceneTiming(audioElements)
	if err != nil {
		s.log.Warnf("Failed to analyze scene timing: %v, using fallback", err)
		sceneTiming = s.generateFallbackTiming(audioElements)
	}

	// Add inputs
	builder.addInput("-y") // Overwrite output
	builder.addInput("-protocol_whitelist", "file,http,https,tcp,tls")
	
	// Background video with loop
	loopsNeeded := int(totalDuration/backgroundVideo.Duration) + 1
	builder.addInput("-stream_loop", fmt.Sprintf("%d", loopsNeeded), "-i", backgroundVideo.Src)

	// Audio inputs
	for _, audio := range audioElements {
		builder.addInput("-i", audio.Src)
	}

	// Image inputs
	for _, image := range imageElements {
		builder.addInput("-i", image.Src)
	}

	// Build filter complex with subtitle support and scene timing
	filterComplex := s.buildFilterComplexWithSubtitlesAndTiming(project, audioElements, imageElements, sceneTiming, totalDuration, subtitleFilePath)

	if filterComplex != "" {
		builder.addArg("-filter_complex", filterComplex)
	}

	// Map outputs
	outputVideoStream := s.getOutputVideoStream(imageElements, subtitleFilePath)
	builder.addArg("-map", outputVideoStream)

	if len(audioElements) > 0 {
		builder.addArg("-map", "[final_audio]")
	}

	// Set duration
	builder.addArg("-t", fmt.Sprintf("%.2f", totalDuration))

	// Output settings based on project config
	s.addOutputSettingsForProject(builder, project)

	// Generate output path
	outputPath := s.generateOutputPathForProject(project)
	builder.addArg(outputPath)

	return &FFmpegCommand{
		Args:       builder.args,
		OutputPath: outputPath,
	}, nil
}


func (s *ffmpegService) addSubtitleFilter(filters *[]string, currentVideo string, subtitleFilePath string) string {
	s.log.Infof("Adding subtitle overlay: %s", subtitleFilePath)
	
	if currentVideo == "0:v" {
		*filters = append(*filters, fmt.Sprintf("[0:v]ass='%s'[subtitled_video]", subtitleFilePath))
	} else {
		*filters = append(*filters, fmt.Sprintf("[%s]ass='%s'[subtitled_video]", currentVideo, subtitleFilePath))
	}
	
	return "subtitled_video"
}

func (s *ffmpegService) analyzeSceneTiming(audioElements []models.Element) ([]models.TimingSegment, error) {
	return s.audio.CalculateSceneTiming(audioElements)
}

func (s *ffmpegService) generateFallbackTiming(audioElements []models.Element) []models.TimingSegment {
	segments := make([]models.TimingSegment, len(audioElements))
	currentTime := 0.0
	
	for i, audio := range audioElements {
		duration := audio.Duration
		if duration <= 0 {
			duration = 5.0 // default fallback
		}
		
		segments[i] = models.TimingSegment{
			StartTime: currentTime,
			EndTime:   currentTime + duration,
			AudioFile: audio.Src,
		}
		currentTime += duration
	}
	
	return segments
}

func (s *ffmpegService) buildFilterComplexWithSubtitlesAndTiming(project models.VideoProject, audioElements, imageElements []models.Element, sceneTiming []models.TimingSegment, totalDuration float64, subtitleFilePath string) string {
	var filters []string
	
	// Audio concatenation
	if len(audioElements) > 1 {
		audioInputs := make([]string, len(audioElements))
		for i := range audioElements {
			audioInputs[i] = fmt.Sprintf("[%d:a]", i+1) // +1 because 0 is background video
		}
		audioConcat := fmt.Sprintf("%sconcat=n=%d:v=0:a=1[concatenated_audio]",
			strings.Join(audioInputs, ""),
			len(audioElements))
		filters = append(filters, audioConcat)
		filters = append(filters, "[concatenated_audio]apad=pad_dur=2[final_audio]")
	} else if len(audioElements) == 1 {
		filters = append(filters, "[1:a]apad=pad_dur=2[final_audio]")
	}
	
	// Image overlays with timing based on actual audio analysis
	currentInput := "0:v"
	
	for i, image := range imageElements {
		// Use scene timing from audio analysis
		var startTime, endTime float64
		if i < len(sceneTiming) {
			startTime = sceneTiming[i].StartTime
			endTime = sceneTiming[i].EndTime
		} else {
			// Fallback if we have more images than timing segments
			startTime = float64(i) * 5.0
			endTime = startTime + 5.0
		}
		
		s.log.Debugf("Image %d overlay timing: %.2fs - %.2fs (duration: %.2fs)", 
			i, startTime, endTime, endTime-startTime)
		
		// Scale image - use correct input index for images with :v selector
		imageInputIndex := len(audioElements) + 1 + i
		scaleFilter := fmt.Sprintf("[%d:v]scale=500:500[scaled_img_%d]",
			imageInputIndex, i)
		filters = append(filters, scaleFilter)
		
		// Overlay with timing based on actual audio duration
		overlayFilter := fmt.Sprintf("[%s][scaled_img_%d]overlay=%d:%d:enable='between(t\\,%f\\,%f)'[overlay_%d]",
			currentInput, i, image.X, image.Y, startTime, endTime, i)
		filters = append(filters, overlayFilter)
		
		currentInput = fmt.Sprintf("overlay_%d", i)
	}
	
	// Add subtitle filter if subtitle file is provided
	if subtitleFilePath != "" {
		finalVideoStream := s.addSubtitleFilter(&filters, currentInput, subtitleFilePath)
		// Update the final output stream name
		_ = finalVideoStream
	}
	
	return strings.Join(filters, ";")
}

func (s *ffmpegService) getOutputVideoStream(imageElements []models.Element, subtitleFilePath string) string {
	if subtitleFilePath != "" {
		return "[subtitled_video]"
	} else if len(imageElements) > 0 {
		return fmt.Sprintf("[overlay_%d]", len(imageElements)-1)
	} else {
		return "0:v"
	}
}