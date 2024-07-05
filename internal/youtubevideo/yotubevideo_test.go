package youtubevideo

import "testing"

func TestExtractYouTubeVideoID(t *testing.T) {
	testCases := []struct {
		url        string
		expectedID string
		shouldErr  bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://youtube.com/v/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://youtube.com/attribution_link?a=dQw4w9WgXcQ&u=/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ", false},
		{"https://www.youtube.com/playlist?list=PL9tY0BWXOZFtgyO1IYJTyWHT1Kt-pQQeB", "", true},
		{"https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw", "", true},
		{"https://www.example.com/watch?v=dQw4w9WgXcQ", "", true},
		{"https://www.youtube.com/watch?foo=bar", "", true},
		{"", "", true},
	}

	for _, tc := range testCases {
		id, err := ExtractYouTubeVideoID(tc.url)
		if tc.shouldErr {
			if err == nil {
				t.Errorf("Expected error for URL: %s, but got none", tc.url)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for URL: %s - %v", tc.url, err)
			}
			if id != tc.expectedID {
				t.Errorf("Expected ID: %s, got: %s for URL: %s", tc.expectedID, id, tc.url)
			}
		}
	}
}

func TestIsYouTubeVideoURL(t *testing.T) {
	testCases := []struct {
		url        string
		isVideoURL bool
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", true},
		{"https://youtu.be/dQw4w9WgXcQ", true},
		{"https://youtube.com/v/dQw4w9WgXcQ", true},
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", true},
		{"https://youtube.com/shorts/dQw4w9WgXcQ", true},
		{"https://youtube.com/attribution_link?a=dQw4w9WgXcQ&u=/watch?v=dQw4w9WgXcQ", true},
		{"https://www.youtube.com/playlist?list=PL9tY0BWXOZFtgyO1IYJTyWHT1Kt-pQQeB", false},
		{"https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw", false},
		{"https://www.example.com/watch?v=dQw4w9WgXcQ", false},
		{"https://www.youtube.com/watch?foo=bar", false},
		{"", false},
	}

	for _, tc := range testCases {
		if result := IsYouTubeVideoURL(tc.url); result != tc.isVideoURL {
			t.Errorf("Expected IsYouTubeVideoURL(%s) to be %v, but got %v", tc.url, tc.isVideoURL, result)
		}
	}
}

func TestIsYouTubePlaylistURL(t *testing.T) {
	testCases := []struct {
		url           string
		isPlaylistURL bool
	}{
		{"https://www.youtube.com/playlist?list=PL9tY0BWXOZFtgyO1IYJTyWHT1Kt-pQQeB", true},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", false},
		{"https://youtu.be/dQw4w9WgXcQ", false},
		{"https://youtube.com/v/dQw4w9WgXcQ", false},
		{"https://www.youtube.com/embed/dQw4w9WgXcQ", false},
		{"https://youtube.com/shorts/dQw4w9WgXcQ", false},
		{"https://youtube.com/attribution_link?a=dQw4w9WgXcQ&u=/watch?v=dQw4w9WgXcQ", false},
		{"https://www.youtube.com/channel/UC_x5XG1OV2P6uZZ5FSM9Ttw", false},
		{"https://www.example.com/watch?v=dQw4w9WgXcQ", false},
		{"https://www.youtube.com/watch?foo=bar", false},
		{"", false},
	}

	for _, tc := range testCases {
		if result := IsYouTubePlaylistURL(tc.url); result != tc.isPlaylistURL {
			t.Errorf("Expected IsYouTubePlaylistURL(%s) to be %v, but got %v", tc.url, tc.isPlaylistURL, result)
		}
	}
}
