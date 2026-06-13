package downloader

import (
	"testing"

	"github.com/kkdai/youtube/v2"
)

func TestSelectFormat(t *testing.T) {
	formats := youtube.FormatList{
		{ItagNo: 18, QualityLabel: "360p", AudioChannels: 2, Bitrate: 500},
		{ItagNo: 22, QualityLabel: "720p", AudioChannels: 2, Bitrate: 1000},
		{ItagNo: 140, AudioChannels: 2, Bitrate: 128000},
		{ItagNo: 251, AudioChannels: 2, Bitrate: 160000},
		{ItagNo: 137, QualityLabel: "1080p", AudioChannels: 0, Bitrate: 2000},
	}

	tests := []struct {
		name string
		req  Request
		want int
	}{
		{name: "best muxed video", req: Request{MediaType: MediaVideo, Quality: QualityBest}, want: 22},
		{name: "worst muxed video", req: Request{MediaType: MediaVideo, Quality: QualityWorst}, want: 18},
		{name: "best audio", req: Request{MediaType: MediaAudio, Quality: QualityBest}, want: 251},
		{name: "worst audio", req: Request{MediaType: MediaAudio, Quality: QualityWorst}, want: 140},
		{name: "exact itag", req: Request{Itag: 137}, want: 137},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectFormat(formats, test.req)
			if err != nil {
				t.Fatalf("selectFormat() error = %v", err)
			}
			if got.ItagNo != test.want {
				t.Fatalf("selectFormat() itag = %d, want %d", got.ItagNo, test.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := map[string]string{
		`  My / video: "test"?  `: "My - video- test",
		`...`:                     "youtube-download",
		`hello   world`:           "hello world",
	}

	for input, want := range tests {
		if got := sanitizeFilename(input); got != want {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtensionFor(t *testing.T) {
	tests := map[string]string{
		"video/mp4; codecs=\"avc1\"":  ".mp4",
		"audio/mp4; codecs=\"mp4a\"":  ".m4a",
		"audio/webm; codecs=\"opus\"": ".webm",
		"invalid":                     ".bin",
	}

	for input, want := range tests {
		if got := extensionFor(input); got != want {
			t.Errorf("extensionFor(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestOutputFilename(t *testing.T) {
	tests := []struct {
		name     string
		request  Request
		title    string
		mimeType string
		want     string
	}{
		{
			name:     "audio uses mp3 extension",
			request:  Request{MediaType: MediaAudio},
			title:    "My song",
			mimeType: "audio/webm",
			want:     "My song.mp3",
		},
		{
			name:     "audio replaces informed extension",
			request:  Request{MediaType: MediaAudio, Filename: "song.m4a"},
			mimeType: "audio/mp4",
			want:     "song.mp3",
		},
		{
			name:     "video keeps informed extension",
			request:  Request{MediaType: MediaVideo, Filename: "video.custom"},
			mimeType: "video/mp4",
			want:     "video.custom",
		},
		{
			name:     "video gets format extension",
			request:  Request{MediaType: MediaVideo, Filename: "video"},
			mimeType: "video/mp4",
			want:     "video.mp4",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := outputFilename(test.request, test.title, test.mimeType)
			if got != test.want {
				t.Fatalf("outputFilename() = %q, want %q", got, test.want)
			}
		})
	}
}
