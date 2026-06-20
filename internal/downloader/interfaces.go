package downloader

import "context"

// AudioConverter converts a downloaded audio stream to MP3.
type AudioConverter interface {
	Convert(ctx context.Context, inputPath, outputPath string, metadata ID3Metadata) error
}
