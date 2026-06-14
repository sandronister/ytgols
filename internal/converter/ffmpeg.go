package converter

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/sandronister/ytgols/internal/downloader"
)

// FFmpeg converts audio files using the ffmpeg executable.
type FFmpeg struct {
	command Command
}

// NewFFmpeg creates an MP3 converter backed by ffmpeg.
func NewFFmpeg() *FFmpeg {
	return &FFmpeg{command: execCommand{}}
}

// Convert converts the input audio file to MP3.
func (f *FFmpeg) Convert(ctx context.Context, inputPath, outputPath string, metadata downloader.ID3Metadata) error {
	args := []string{
		"-nostdin",
		"-y",
		"-i", inputPath,
		"-vn",
		"-codec:a", "libmp3lame",
		"-q:a", "2",
		"-id3v2_version", "4",
	}
	args = appendMetadata(args, metadata)
	args = append(args, outputPath)

	output, err := f.command.Run(ctx, "ffmpeg", args...)
	if err == nil {
		return nil
	}

	if errors.Is(err, exec.ErrNotFound) {
		return errors.New("ffmpeg não está instalado ou não foi encontrado no PATH")
	}
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Errorf("executar ffmpeg: %w", err)
	}
	return fmt.Errorf("executar ffmpeg: %w: %s", err, message)
}

func appendMetadata(args []string, metadata downloader.ID3Metadata) []string {
	fields := []struct {
		key   string
		value string
	}{
		{key: "title", value: metadata.Title},
		{key: "artist", value: metadata.Artist},
		{key: "album", value: metadata.Album},
		{key: "date", value: metadata.Year},
		{key: "genre", value: metadata.Genre},
		{key: "track", value: metadata.Track},
	}
	for _, field := range fields {
		if field.value != "" {
			args = append(args, "-metadata", field.key+"="+field.value)
		}
	}
	return args
}

type execCommand struct{}

func (execCommand) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
