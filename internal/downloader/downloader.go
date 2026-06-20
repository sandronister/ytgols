// Package downloader implements YouTube media downloads for the application.
package downloader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	written, err := d.downloadFormat(ctx, video, format, file, req.Progress)
	if err != nil {
		return Result{}, err
	}
	closeErr := file.Close()
	fileClosed = true
	if closeErr != nil {
		return Result{}, fmt.Errorf("close output file: %w", closeErr)
	}
	completed = true

	if req.MediaType == MediaAudio {
		if err := d.audioConverter.Convert(ctx, downloadPath, path); err != nil {
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

func (d *Downloader) downloadFormat(
	ctx context.Context,
	video *youtube.Video,
	format youtube.Format,
	file *os.File,
	progress func(downloaded, total int64),
) (int64, error) {
	written, err := d.copyFormat(ctx, video, format, file, progress)
	if err == nil {
		return written, nil
	}
	if !isUnexpectedStatus(err, http.StatusForbidden) {
		return 0, err
	}

	if _, seekErr := file.Seek(0, io.SeekStart); seekErr != nil {
		return 0, fmt.Errorf("prepare download retry: %w", seekErr)
	}
	if truncateErr := file.Truncate(0); truncateErr != nil {
		return 0, fmt.Errorf("prepare download retry: %w", truncateErr)
	}

	fallbackFormat := format
	fallbackFormat.ContentLength = 0
	written, retryErr := d.copyFormat(ctx, video, fallbackFormat, file, progress)
	if retryErr == nil {
		return written, nil
	}

	return 0, fmt.Errorf(
		"download media: YouTube returned HTTP 403 for this stream and the non-chunked retry also failed: %w. Try another itag/quality or update the downloader library if YouTube changed its stream rules",
		retryErr,
	)
}

func (d *Downloader) copyFormat(
	ctx context.Context,
	video *youtube.Video,
	format youtube.Format,
	file *os.File,
	progress func(downloaded, total int64),
) (written int64, err error) {
	stream, size, err := d.client.GetStreamContext(ctx, video, &format)
	if err != nil {
		return 0, fmt.Errorf("open media stream: %w", err)
	}

	writer := io.Writer(file)
	if progress != nil {
		writer = &progressWriter{
			writer: writer,
			total:  size,
			report: progress,
		}
	}

	written, err = io.Copy(writer, stream)
	closeErr := stream.Close()
	if err != nil {
		return written, fmt.Errorf("download media: %w", err)
	}
	if closeErr != nil {
		return written, fmt.Errorf("close media stream: %w", closeErr)
	}

	return written, nil
}

func isUnexpectedStatus(err error, status int) bool {
	var statusErr youtube.ErrUnexpectedStatusCode
	return errors.As(err, &statusErr) && int(statusErr) == status
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
