package downloader

import (
	"testing"

	"github.com/kkdai/youtube/v2"
)

func TestMetadataFromVideoUsesYouTubeMetadata(t *testing.T) {
	got := metadataFromVideo(ID3Metadata{}, &youtube.Video{
		Title:  "Song",
		Author: "Artist - Topic",
	})

	if got.Title != "Song" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Artist != "Artist" {
		t.Fatalf("Artist = %q", got.Artist)
	}
}

func TestMetadataFromVideoSplitsTitleWhenAuthorIsEmpty(t *testing.T) {
	got := metadataFromVideo(ID3Metadata{}, &youtube.Video{
		Title: "Artist - Song",
	})

	if got.Title != "Song" {
		t.Fatalf("Title = %q", got.Title)
	}
	if got.Artist != "Artist" {
		t.Fatalf("Artist = %q", got.Artist)
	}
}

func TestMetadataFromVideoKeepsExplicitMetadata(t *testing.T) {
	got := metadataFromVideo(ID3Metadata{
		Title:  "Custom song",
		Artist: "Custom artist",
		Album:  "Custom album",
	}, &youtube.Video{
		Title:  "Song",
		Author: "Artist",
	})

	if got.Title != "Custom song" || got.Artist != "Custom artist" || got.Album != "Custom album" {
		t.Fatalf("metadataFromVideo() = %#v", got)
	}
}
