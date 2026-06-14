package form

// Answers contains the values collected by the interactive form.
type Answers struct {
	URL       string
	MediaType string
	Quality   string
	OutputDir string
	Filename  string
	Itag      int
	Metadata  ID3Metadata
}

// ID3Metadata contains optional metadata for audio downloads.
type ID3Metadata struct {
	Title  string
	Artist string
	Album  string
	Year   string
	Genre  string
	Track  string
}
