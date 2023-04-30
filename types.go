package main

// What we get from the default config file
type ConfigData struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Category      string   `json:"category"`
	PrivacyStatus string   `json:"privacy_status"`
}

// What we need inside the form
type FormData struct {
	Title              string
	Description        string
	Tags               []string
	CategoryIndex      int
	PrivacyStatusIndex int
}

// What we need to send to YouTube to upload a video
type VideoUploadData struct {
	VideoPath     string
	ThumbnailPath string
	Title         string
	Description   string
	Tags          []string
	Category      string
	PrivacyStatus string
	PublishAt     string
}

// What YouTube gives us when we list latest videos
type VideoDownloadData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoId     string `json:"video_id"`
}
