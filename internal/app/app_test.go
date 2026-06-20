package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	media "github.com/sandronister/ytgols/internal/downloader"
	"github.com/sandronister/ytgols/internal/form"
)

type formStub struct {
	answers form.Answers
	err     error
}

func (s formStub) Ask() (form.Answers, error) {
	return s.answers, s.err
}

type downloaderStub struct {
	request media.Request
	result  media.Result
	err     error
}

func (s *downloaderStub) Download(_ context.Context, request media.Request) (media.Result, error) {
	s.request = request
	if request.Progress != nil {
		request.Progress(50, 100)
		request.Progress(50, 100)
		request.Progress(100, 100)
	}
	return s.result, s.err
}

func TestRun(t *testing.T) {
	downloadService := &downloaderStub{
		result: media.Result{Path: "/tmp/video.mp4"},
	}
	var progress bytes.Buffer
	application := New(formStub{
		answers: form.Answers{
			URL:       "https://youtu.be/example",
			MediaType: "audio",
			Quality:   "worst",
			OutputDir: "downloads",
			Filename:  "example",
			Itag:      140,
		},
	}, downloadService, &progress)
	workingDirectory := t.TempDir()
	application.workingDirectory = func() (string, error) {
		return workingDirectory, nil
	}

	result, err := application.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Path != "/tmp/video.mp4" {
		t.Fatalf("result.Path = %q", result.Path)
	}
	if downloadService.request.URL != "https://youtu.be/example" {
		t.Errorf("request.URL = %q", downloadService.request.URL)
	}
	if downloadService.request.OutputDir != filepath.Join(filepath.Dir(workingDirectory), "downloads") {
		t.Errorf("request.OutputDir = %q", downloadService.request.OutputDir)
	}
	if downloadService.request.MediaType != media.MediaAudio {
		t.Errorf("request.MediaType = %q", downloadService.request.MediaType)
	}
	if downloadService.request.Quality != media.QualityWorst {
		t.Errorf("request.Quality = %q", downloadService.request.Quality)
	}
	if downloadService.request.Itag != 140 {
		t.Errorf("request.Itag = %d", downloadService.request.Itag)
	}
	if count := strings.Count(progress.String(), "50%"); count != 1 {
		t.Errorf("50%% printed %d times", count)
	}
	if !strings.Contains(progress.String(), "100%") {
		t.Error("100% progress was not printed")
	}
}

func TestDownloadDirectory(t *testing.T) {
	workingDirectory := t.TempDir()
	existingDirectory := filepath.Join(workingDirectory, "downloads")
	absoluteDirectory := filepath.Join(filepath.VolumeName(workingDirectory)+string(filepath.Separator), "tmp", "downloads")
	if err := os.Mkdir(existingDirectory, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		directory string
		working   string
		want      string
	}{
		{
			name:      "existing directory in execution location",
			directory: "downloads",
			working:   workingDirectory,
			want:      existingDirectory,
		},
		{
			name:      "new directory in parent",
			directory: filepath.Join("videos", "musicas"),
			working:   workingDirectory,
			want:      filepath.Join(filepath.Dir(workingDirectory), "videos", "musicas"),
		},
		{
			name:      "absolute directory",
			directory: absoluteDirectory,
			working:   workingDirectory,
			want:      absoluteDirectory,
		},
		{
			name:      "execution at filesystem root",
			directory: "downloads",
			working:   string(filepath.Separator),
			want:      filepath.Join(string(filepath.Separator), "downloads"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := downloadDirectory(test.directory, test.working)
			if err != nil {
				t.Fatalf("downloadDirectory() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("downloadDirectory() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDownloadDirectoryRejectsFile(t *testing.T) {
	workingDirectory := t.TempDir()
	path := filepath.Join(workingDirectory, "downloads")
	if err := os.WriteFile(path, []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := downloadDirectory("downloads", workingDirectory)
	if err == nil || !strings.Contains(err.Error(), "não é um diretório") {
		t.Fatalf("downloadDirectory() error = %v", err)
	}
}

func TestRunFormError(t *testing.T) {
	application := New(
		formStub{err: errors.New("input closed")},
		&downloaderStub{},
		nil,
	)

	_, err := application.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "ler informações") {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestProgressWithoutTotal(t *testing.T) {
	var output bytes.Buffer
	reporter := newProgressReporter(&output)

	reporter.Report(1024*1024, 0)

	if got := output.String(); !strings.Contains(got, "1.0 MiB") {
		t.Fatalf("progress output = %q", got)
	}
}
