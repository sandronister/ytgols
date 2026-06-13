package converter

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
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
func (f *FFmpeg) Convert(ctx context.Context, inputPath, outputPath string) error {
	output, err := f.command.Run(
		ctx,
		"ffmpeg",
		"-nostdin",
		"-y",
		"-i", inputPath,
		"-vn",
		"-codec:a", "libmp3lame",
		"-q:a", "2",
		outputPath,
	)
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

type execCommand struct{}

func (execCommand) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}
