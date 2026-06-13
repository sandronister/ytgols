package form

import (
	"bytes"
	"strings"
	"testing"
)

func TestAskWithDefaults(t *testing.T) {
	input := strings.NewReader("https://youtu.be/example\n\n\n\n\n\n")
	var output bytes.Buffer

	got, err := New(input, &output).Ask()
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}

	if got.URL != "https://youtu.be/example" {
		t.Errorf("URL = %q", got.URL)
	}
	if got.MediaType != "video" {
		t.Errorf("MediaType = %q", got.MediaType)
	}
	if got.Quality != "best" {
		t.Errorf("Quality = %q", got.Quality)
	}
	if got.OutputDir != "downloads" {
		t.Errorf("OutputDir = %q", got.OutputDir)
	}
	if got.Filename != "" {
		t.Errorf("Filename = %q", got.Filename)
	}
	if got.Itag != 0 {
		t.Errorf("Itag = %d", got.Itag)
	}
}

func TestAskWithValuesAndValidation(t *testing.T) {
	input := strings.NewReader(strings.Join([]string{
		"",
		"https://youtube.com/watch?v=example",
		"invalid",
		"AUDIO",
		"worst",
		"meus downloads",
		"minha musica",
		"-1",
		"140",
		"",
	}, "\n"))
	var output bytes.Buffer

	got, err := New(input, &output).Ask()
	if err != nil {
		t.Fatalf("Ask() error = %v", err)
	}

	if got.MediaType != "audio" || got.Quality != "worst" {
		t.Fatalf("MediaType/Quality = %q/%q", got.MediaType, got.Quality)
	}
	if got.OutputDir != "meus downloads" || got.Filename != "minha musica" {
		t.Fatalf("OutputDir/Filename = %q/%q", got.OutputDir, got.Filename)
	}
	if got.Itag != 140 {
		t.Fatalf("Itag = %d", got.Itag)
	}
	if !strings.Contains(output.String(), "Valor inválido") {
		t.Error("expected validation message")
	}
}
