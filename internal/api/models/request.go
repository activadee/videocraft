package models

import (
	"errors"
	"time"
)

type VideoConfigArray []VideoProject

type VideoProject struct {
	Comment    string    `json:"comment,omitempty"`
	Resolution string    `json:"resolution,omitempty"`
	Quality    string    `json:"quality,omitempty"`
	Width      int       `json:"width,omitempty"`
	Height     int       `json:"height,omitempty"`
	Scenes     []Scene   `json:"scenes,omitempty"`
	Elements   []Element `json:"elements,omitempty"`
}

type Scene struct {
	ID              string    `json:"id"`
	BackgroundColor string    `json:"background-color,omitempty"`
	Elements        []Element `json:"elements,omitempty"`
}

type Element struct {
	Type string `json:"type"`
	Src  string `json:"src,omitempty"`
	ID   string `json:"id,omitempty"`

	X int `json:"x,omitempty"`
	Y int `json:"y,omitempty"`

	ZIndex   int     `json:"z-index,omitempty"`
	Volume   float64 `json:"volume,omitempty"`
	Resize   string  `json:"resize,omitempty"`
	Duration float64 `json:"duration,omitempty"`

	Settings SubtitleSettings `json:"settings,omitempty"`
	Language string           `json:"language,omitempty"`
}

type SubtitleSettings struct {
	Style        string `json:"style,omitempty"`
	FontFamily   string `json:"font-family,omitempty"`
	FontSize     int    `json:"font-size,omitempty"`
	WordColor    string `json:"word-color,omitempty"`
	LineColor    string `json:"line-color,omitempty"`
	ShadowColor  string `json:"shadow-color,omitempty"`
	ShadowOffset int    `json:"shadow-offset,omitempty"`
	BoxColor     string `json:"box-color,omitempty"`
	Position     string `json:"position,omitempty"`
	OutlineColor string `json:"outline-color,omitempty"`
	OutlineWidth int    `json:"outline-width,omitempty"`
}

// Validation
func (vca VideoConfigArray) Validate() error {
	if len(vca) == 0 {
		return errors.New("at least one video project is required")
	}

	for i, project := range vca {
		if err := project.Validate(); err != nil {
			return errors.New("project " + string(rune(i)) + ": " + err.Error())
		}
	}

	return nil
}

func (vp VideoProject) Validate() error {
	// Validate scenes
	for i, scene := range vp.Scenes {
		if scene.ID == "" {
			return errors.New("scene " + string(rune(i)) + ": ID is required")
		}

		for j, element := range scene.Elements {
			if err := element.Validate(); err != nil {
				return errors.New("scene " + scene.ID + " element " + string(rune(j)) + ": " + err.Error())
			}
		}
	}

	// Validate global elements
	for i, element := range vp.Elements {
		if err := element.Validate(); err != nil {
			return errors.New("global element " + string(rune(i)) + ": " + err.Error())
		}
	}

	return nil
}

func (e Element) Validate() error {
	if e.Type == "" {
		return errors.New("element type is required")
	}

	// Validate based on type
	switch e.Type {
	case "video", "audio", "image":
		if e.Src == "" {
			return errors.New("src is required for " + e.Type + " elements")
		}
	case "subtitles":
		// Subtitles don't require src
	default:
		return errors.New("unsupported element type: " + e.Type)
	}

	if e.Duration < 0 {
		return errors.New("duration cannot be negative")
	}

	return nil
}

// TimingSegment represents a timing segment for video generation
type TimingSegment struct {
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	AudioFile  string  `json:"audio_file"`
	Text       string  `json:"text,omitempty"`
	Transcript string  `json:"transcript,omitempty"`
}

// Job model
type Job struct {
	ID          string           `json:"id"`
	Status      JobStatus        `json:"status"`
	Config      VideoConfigArray `json:"config"`
	VideoID     string           `json:"video_id,omitempty"`
	Error       string           `json:"error,omitempty"`
	Progress    int              `json:"progress"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// VideoInfo contains comprehensive video file metadata
type VideoInfo struct {
	ID        string  `json:"id"`
	Filename  string  `json:"filename"`
	Size      int64   `json:"size"`
	CreatedAt string  `json:"created_at"`
	URL       string  `json:"url,omitempty"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	Duration  float64 `json:"duration"`
	Format    string  `json:"format"`
	Codec     string  `json:"codec,omitempty"`
}

// GetDuration returns the video duration - implements common interface for job service
func (vi *VideoInfo) GetDuration() float64 {
	return vi.Duration
}

// ImageInfo contains comprehensive image file metadata
type ImageInfo struct {
	ID            string `json:"id,omitempty"`
	Filename      string `json:"filename,omitempty"`
	URL           string `json:"url,omitempty"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Format        string `json:"format"`
	Size          int64  `json:"size"`
	Path          string `json:"path,omitempty"`
	ProcessedPath string `json:"processed_path,omitempty"`
}
