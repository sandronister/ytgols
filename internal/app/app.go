package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandronister/ytgols/internal/downloader"
)

const progressLineWidth = 40

// App coordinates the interactive form and the media download.
type App struct {
	form             Form
	downloader       Downloader
	progress         io.Writer
	workingDirectory func() (string, error)
}

// New creates the command-line application.
func New(form Form, downloader Downloader, progress io.Writer) *App {
	return &App{
		form:             form,
		downloader:       downloader,
		progress:         progress,
		workingDirectory: os.Getwd,
	}
}

// Run collects the settings, downloads the media and returns its result.
func (a *App) Run(ctx context.Context) (downloader.Result, error) {
	answers, err := a.form.Ask()
	if err != nil {
		return downloader.Result{}, fmt.Errorf("ler informações: %w", err)
	}

	workingDirectory, err := a.workingDirectory()
	if err != nil {
		return downloader.Result{}, fmt.Errorf("identificar diretório de execução: %w", err)
	}
	outputDirectory, err := downloadDirectory(answers.OutputDir, workingDirectory)
	if err != nil {
		return downloader.Result{}, err
	}

	reporter := newProgressReporter(a.progress)
	result, err := a.downloader.Download(ctx, downloader.Request{
		URL:       answers.URL,
		OutputDir: outputDirectory,
		Filename:  answers.Filename,
		MediaType: downloader.MediaType(answers.MediaType),
		Quality:   downloader.Quality(answers.Quality),
		Itag:      answers.Itag,
		Progress:  reporter.Report,
	})
	if err != nil {
		return downloader.Result{}, err
	}

	reporter.Clear()
	return result, nil
}

func downloadDirectory(directory, workingDirectory string) (string, error) {
	if filepath.IsAbs(directory) {
		return filepath.Clean(directory), nil
	}

	currentLocation := filepath.Join(workingDirectory, directory)
	info, err := os.Stat(currentLocation)
	switch {
	case err == nil && info.IsDir():
		return filepath.Clean(currentLocation), nil
	case err == nil:
		return "", fmt.Errorf("destino %q existe e não é um diretório", currentLocation)
	case !os.IsNotExist(err):
		return "", fmt.Errorf("verificar diretório de destino: %w", err)
	}

	parent := filepath.Dir(workingDirectory)
	if parent != workingDirectory {
		return filepath.Clean(filepath.Join(parent, directory)), nil
	}
	return filepath.Clean(currentLocation), nil
}

type progressReporter struct {
	output      io.Writer
	lastPercent int64
}

func newProgressReporter(output io.Writer) *progressReporter {
	return &progressReporter{
		output:      output,
		lastPercent: -1,
	}
}

func (r *progressReporter) Report(downloaded, total int64) {
	if r.output == nil {
		return
	}
	if total <= 0 {
		fmt.Fprintf(r.output, "\rBaixado: %.1f MiB", float64(downloaded)/(1024*1024))
		return
	}

	percent := downloaded * 100 / total
	if percent == r.lastPercent {
		return
	}
	fmt.Fprintf(r.output, "\rProgresso: %3d%%", percent)
	r.lastPercent = percent
}

func (r *progressReporter) Clear() {
	if r.output != nil {
		fmt.Fprintf(r.output, "\r%s\r", strings.Repeat(" ", progressLineWidth))
	}
}
