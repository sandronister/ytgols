package app

import (
	"context"

	"github.com/sandronister/ytgols/internal/downloader"
	"github.com/sandronister/ytgols/internal/form"
)

// Form collects the download settings.
type Form interface {
	Ask() (form.Answers, error)
}

// Downloader downloads media based on a request.
type Downloader interface {
	Download(context.Context, downloader.Request) (downloader.Result, error)
}
