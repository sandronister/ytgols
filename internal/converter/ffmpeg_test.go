package converter

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

type commandStub struct {
	name   string
	args   []string
	output []byte
	err    error
}

func (s *commandStub) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	s.name = name
	s.args = args
	return s.output, s.err
}

func TestConvert(t *testing.T) {
	command := &commandStub{}
	ffmpeg := &FFmpeg{command: command}

	err := ffmpeg.Convert(context.Background(), "input.webm", "output.mp3")
	if err != nil {
		t.Fatalf("Convert() error = %v", err)
	}

	wantArgs := []string{
		"-nostdin", "-y",
		"-i", "input.webm",
		"-vn",
		"-codec:a", "libmp3lame",
		"-q:a", "2",
		"output.mp3",
	}
	if command.name != "ffmpeg" {
		t.Errorf("command name = %q", command.name)
	}
	if !reflect.DeepEqual(command.args, wantArgs) {
		t.Errorf("command args = %#v, want %#v", command.args, wantArgs)
	}
}

func TestConvertMissingFFmpeg(t *testing.T) {
	ffmpeg := &FFmpeg{command: &commandStub{err: exec.ErrNotFound}}

	err := ffmpeg.Convert(context.Background(), "input.webm", "output.mp3")
	if err == nil || !strings.Contains(err.Error(), "não está instalado") {
		t.Fatalf("Convert() error = %v", err)
	}
}

func TestConvertIncludesFFmpegOutput(t *testing.T) {
	ffmpeg := &FFmpeg{command: &commandStub{
		output: []byte("invalid input"),
		err:    errors.New("exit status 1"),
	}}

	err := ffmpeg.Convert(context.Background(), "input.webm", "output.mp3")
	if err == nil || !strings.Contains(err.Error(), "invalid input") {
		t.Fatalf("Convert() error = %v", err)
	}
}
