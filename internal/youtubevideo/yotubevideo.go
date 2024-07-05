package youtubevideo

import (
	"fmt"
	"github.com/kkdai/youtube/v2"
	"regexp"
)

var (
	videoUrlRegex    = regexp.MustCompile(`(?:https?://)?(?:www\.)?(?:youtube\.com/(?:.*[?&]v=|(?:v|embed|shorts)/|watch\?.*?v=)|youtu\.be/)([a-zA-Z0-9_-]{11})`)
	playlistUrlRegex = regexp.MustCompile(`(?:https?://)?(?:www\.)?youtube\.com/playlist\?list=([a-zA-Z0-9_-]+)`)
)

func ExtractYouTubeVideoID(url string) (string, error) {
	matches := videoUrlRegex.FindStringSubmatch(url)

	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("invalid YouTube URL: %s", url)
}

func IsYouTubeVideoURL(url string) bool {
	return videoUrlRegex.MatchString(url)
}

func IsYouTubePlaylistURL(url string) bool {
	return playlistUrlRegex.MatchString(url)
}

func HighestQualityFormat(formats []youtube.Format) *youtube.Format {
	if len(formats) == 0 {
		return nil
	}

	var bestFormat *youtube.Format
	for i, format := range formats {
		if bestFormat == nil || isHigherQuality(format, *bestFormat) {
			bestFormat = &formats[i]
		}
	}
	return bestFormat
}

func isHigherQuality(f1, f2 youtube.Format) bool {
	// Define your criteria for quality comparison. Example: based on resolution and bitrate.
	if f1.Width > f2.Width || (f1.Width == f2.Width && f1.Height > f2.Height) {
		return true
	}
	if f1.Width == f2.Width && f1.Height == f2.Height && f1.Bitrate > f2.Bitrate {
		return true
	}
	return false
}
