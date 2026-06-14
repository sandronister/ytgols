// Package downloader implements YouTube media downloads for the application.
package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkdai/youtube/v2"
)

// Downloader downloads media from YouTube.
type Downloader struct {
	client         *youtube.Client
	audioConverter AudioConverter
}

// New creates a Downloader with the default YouTube client.
func New(audioConverter AudioConverter) *Downloader {
	return &Downloader{
		client:         &youtube.Client{},
		audioConverter: audioConverter,
	}
}

// Formats returns the streams available for a video.
func (d *Downloader) Formats(ctx context.Context, rawURL string) ([]Format, error) {
	video, err := d.client.GetVideoContext(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("get video metadata: %w", err)
	}

	formats := make([]Format, 0, len(video.Formats))
	for _, format := range video.Formats {
		formats = append(formats, newFormat(format))
	}
	return formats, nil
}

// Download selects a stream and saves it to disk.
func (d *Downloader) Download(ctx context.Context, req Request) (result Result, err error) {
	req, err = normalizeRequest(req)
	if err != nil {
		return Result{}, err
	}

	video, err := d.client.GetVideoContext(ctx, req.URL)
	if err != nil {
		return Result{}, fmt.Errorf("get video metadata: %w", err)
	}

	format, err := selectFormat(video.Formats, req)
	if err != nil {
		return Result{}, err
	}

	stream, size, err := d.client.GetStreamContext(ctx, video, &format)
	if err != nil {
		return Result{}, fmt.Errorf("open media stream: %w", err)
	}
	defer func() {
		closeErr := stream.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close media stream: %w", closeErr)
		}
	}()

	outputDir, err := filepath.Abs(req.OutputDir)
	if err != nil {
		return Result{}, fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	filename := outputFilename(req, video.Title, format.MimeType)
	path := filepath.Join(outputDir, filename)
	downloadPath := path
	if req.MediaType == MediaAudio {
		if d.audioConverter == nil {
			return Result{}, errors.New("audio converter is required")
		}
		tempFile, createErr := os.CreateTemp(outputDir, ".ytgols-*"+extensionFor(format.MimeType))
		if createErr != nil {
			return Result{}, fmt.Errorf("create temporary audio file: %w", createErr)
		}
		downloadPath = tempFile.Name()
		if closeErr := tempFile.Close(); closeErr != nil {
			_ = os.Remove(downloadPath)
			return Result{}, fmt.Errorf("close temporary audio file: %w", closeErr)
		}
		defer os.Remove(downloadPath)
	}

	file, err := os.Create(downloadPath)
	if err != nil {
		return Result{}, fmt.Errorf("create output file: %w", err)
	}

	completed := false
	fileClosed := false
	defer func() {
		if !fileClosed {
			closeErr := file.Close()
			if err == nil && closeErr != nil {
				err = fmt.Errorf("close output file: %w", closeErr)
			}
		}
		if !completed {
			_ = os.Remove(downloadPath)
		}
	}()

	writer := io.Writer(file)
	if req.Progress != nil {
		writer = &progressWriter{
			writer: writer,
			total:  size,
			report: req.Progress,
		}
	}

	written, err := io.Copy(writer, stream)
	if err != nil {
		return Result{}, fmt.Errorf("download media: %w", err)
	}
	closeErr := file.Close()
	fileClosed = true
	if closeErr != nil {
		return Result{}, fmt.Errorf("close output file: %w", closeErr)
	}
	completed = true

	if req.MediaType == MediaAudio {
		if err := d.audioConverter.Convert(ctx, downloadPath, path, req.Metadata); err != nil {
			_ = os.Remove(path)
			return Result{}, fmt.Errorf("converter áudio para MP3: %w", err)
		}
		info, statErr := os.Stat(path)
		if statErr != nil {
			return Result{}, fmt.Errorf("read converted audio: %w", statErr)
		}
		written = info.Size()
	}

	return Result{
		Path:   path,
		Title:  video.Title,
		Author: video.Author,
		Format: newFormat(format),
		Bytes:  written,
	}, nil
}

func outputFilename(req Request, title, mimeType string) string {
	filename := req.Filename
	if filename == "" {
		filename = sanitizeFilename(title)
	}
	filename = filepath.Base(filename)
	if req.MediaType == MediaAudio {
		return strings.TrimSuffix(filename, filepath.Ext(filename)) + ".mp3"
	}
	if filepath.Ext(filename) == "" {
		filename += extensionFor(mimeType)
	}
	return filename
}

func normalizeRequest(req Request) (Request, error) {
	if strings.TrimSpace(req.URL) == "" {
		return Request{}, errors.New("video URL is required")
	}
	if req.OutputDir == "" {
		req.OutputDir = "."
	}
	if req.MediaType == "" {
		req.MediaType = MediaVideo
	}
	if req.Quality == "" {
		req.Quality = QualityBest
	}
	if req.MediaType != MediaVideo && req.MediaType != MediaAudio {
		return Request{}, fmt.Errorf("unsupported media type %q", req.MediaType)
	}
	if req.Quality != QualityBest && req.Quality != QualityWorst {
		return Request{}, fmt.Errorf("unsupported quality %q", req.Quality)
	}
	return req, nil
}

type progressWriter struct {
	writer     io.Writer
	total      int64
	downloaded int64
	report     func(downloaded, total int64)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	w.downloaded += int64(n)
	w.report(w.downloaded, w.total)
	return n, err
}
