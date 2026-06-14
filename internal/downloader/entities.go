package downloader

// MediaType determines which kind of YouTube stream is selected.
type MediaType string

const (
	MediaVideo MediaType = "video"
	MediaAudio MediaType = "audio"
)

// Quality controls automatic format selection.
type Quality string

const (
	QualityBest  Quality = "best"
	QualityWorst Quality = "worst"
)

// ID3Metadata contains optional ID3v2.4 fields for an MP3 file.
type ID3Metadata struct {
	Title  string
	Artist string
	Album  string
	Year   string
	Genre  string
	Track  string
}

// Request describes one download.
type Request struct {
	URL       string
	OutputDir string
	Filename  string
	MediaType MediaType
	Quality   Quality
	Itag      int
	Metadata  ID3Metadata
	Progress  func(downloaded, total int64)
}

// Result contains information about a completed download.
type Result struct {
	Path   string
	Title  string
	Author string
	Format Format
	Bytes  int64
}

// Format is a stable representation of a YouTube media stream.
type Format struct {
	Itag          int
	MimeType      string
	Quality       string
	QualityLabel  string
	Bitrate       int
	AudioChannels int
	ContentLength int64
}
