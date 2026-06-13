package downloader

import (
	"fmt"
	"mime"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func newFormat(format youtube.Format) Format {
	return Format{
		Itag:          format.ItagNo,
		MimeType:      format.MimeType,
		Quality:       format.Quality,
		QualityLabel:  format.QualityLabel,
		Bitrate:       format.Bitrate,
		AudioChannels: format.AudioChannels,
		ContentLength: format.ContentLength,
	}
}

func selectFormat(formats youtube.FormatList, req Request) (youtube.Format, error) {
	if req.Itag != 0 {
		for _, format := range formats {
			if format.ItagNo == req.Itag {
				return format, nil
			}
		}
		return youtube.Format{}, fmt.Errorf("format itag %d is not available", req.Itag)
	}

	candidates := make(youtube.FormatList, 0, len(formats))
	for _, format := range formats {
		switch req.MediaType {
		case MediaAudio:
			if format.AudioChannels > 0 && format.QualityLabel == "" {
				candidates = append(candidates, format)
			}
		case MediaVideo:
			if format.AudioChannels > 0 && format.QualityLabel != "" {
				candidates = append(candidates, format)
			}
		}
	}
	if len(candidates) == 0 {
		return youtube.Format{}, fmt.Errorf("no %s format is available", req.MediaType)
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if req.MediaType == MediaVideo {
			left, right := videoHeight(candidates[i]), videoHeight(candidates[j])
			if left != right {
				return left > right
			}
		}
		return candidates[i].Bitrate > candidates[j].Bitrate
	})
	if req.Quality == QualityWorst {
		return candidates[len(candidates)-1], nil
	}
	return candidates[0], nil
}

func videoHeight(format youtube.Format) int {
	var height int
	_, _ = fmt.Sscanf(format.QualityLabel, "%dp", &height)
	return height
}

func extensionFor(mimeType string) string {
	mediaType, _, err := mime.ParseMediaType(mimeType)
	if err == nil {
		switch mediaType {
		case "video/mp4":
			return ".mp4"
		case "audio/mp4":
			return ".m4a"
		case "video/webm", "audio/webm":
			return ".webm"
		}
		if extensions, extensionErr := mime.ExtensionsByType(mediaType); extensionErr == nil && len(extensions) > 0 {
			return extensions[0]
		}
	}
	return ".bin"
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		`"`, "",
		"<", "",
		">", "",
		"|", "-",
	)
	name = strings.Join(strings.Fields(replacer.Replace(name)), " ")
	name = strings.Trim(name, ". ")
	if name == "" || name == "." || name == string(filepath.Separator) {
		return "youtube-download"
	}
	return name
}
