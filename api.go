package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

// YouTubeAPI is the class we use to interact with the YouTube API.
type YouTubeAPI struct {
	service *youtube.Service
}

// NewYouTubeAPI creates a new instance of the YouTubeAPI
// class and initializes the service.
func NewYouTubeAPI() (*YouTubeAPI, error) {
	yt := new(YouTubeAPI)
	service, err := yt.getService()
	if err != nil {
		return nil, err
	}
	yt.service = service
	return yt, nil
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func (yt *YouTubeAPI) getClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	cacheFile, err := yt.tokenCacheFile()
	if err != nil {
		return nil, err
	}
	tok, err := yt.tokenFromFile(cacheFile)
	if err != nil {
		tok, err = yt.getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err = yt.saveToken(cacheFile, tok)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func (yt *YouTubeAPI) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func (yt *YouTubeAPI) tokenCacheFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(homeDir, ".config", "ytup")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir, url.QueryEscape("youtube_api_credentials.json")), nil
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func (yt *YouTubeAPI) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// saveToken uses a file path to create a file and store the
// token in it.
func (yt *YouTubeAPI) saveToken(file string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return err
	}
	return nil
}

// getService returns the service that is needed to interact
// with the YouTube API.
func (yt *YouTubeAPI) getService() (*youtube.Service, error) {
	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadFile(filepath.Join(homeDir, ".config", "ytup", "client_secret.json"))
	if err != nil {
		return nil, err
	}
	config, err := google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
	if err != nil {
		return nil, err
	}
	client, err := yt.getClient(ctx, config)
	if err != nil {
		return nil, err
	}
	service, err := youtube.New(client)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetExtraVideoData returns the tags and category id of a video
// given its id. These are not available when we list the videos.
func (yt *YouTubeAPI) GetExtraVideoData(id string) (*ExtraVideoData, error) {
	listCall := yt.service.Videos.List([]string{"snippet"})
	listResponse, err := listCall.Id(id).Do()
	if err != nil {
		return nil, err
	}

	if len(listResponse.Items) == 0 {
		return nil, errors.New("Video not found")
	}

	tags := listResponse.Items[0].Snippet.Tags
	category := categories[0]
	categoryId, err := strconv.Atoi(listResponse.Items[0].Snippet.CategoryId)
	if err == nil {
		for k, v := range categoryNameToId {
			if v == categoryId {
				category = k
			}
		}
	}
	extraVideoData := ExtraVideoData{
		Tags:     tags,
		Category: category,
	}

	return &extraVideoData, nil
}

// GetLatestVideos returns a slice of data from the 10 most recently
// uploaded videos.
func (yt *YouTubeAPI) GetLatestVideos(readCache bool) ([]*VideoDownloadData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	cacheFilePath := filepath.Join(homeDir, ".config", "ytup", "videos_cache.json")
	_, err = os.Stat(cacheFilePath)
	if os.IsNotExist(err) {
		os.Create(cacheFilePath)
	} else if err != nil {
		return nil, err
	}

	videosCacheFile, err := os.OpenFile(cacheFilePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer videosCacheFile.Close()

	var latestVideos []*VideoDownloadData
	if readCache {
		json.NewDecoder(videosCacheFile).Decode(&latestVideos)
	} else {
		listCall := yt.service.Search.List([]string{"snippet"})
		listResponse, err := listCall.ForMine(true).MaxResults(10).Order("date").Type("video").Do()
		if err != nil {
			return nil, err
		}

		for _, video := range listResponse.Items {
			latestVideos = append(latestVideos, &VideoDownloadData{
				Title:       video.Snippet.Title,
				Description: video.Snippet.Description,
				VideoId:     video.Id.VideoId,
			})
		}
		json.NewEncoder(videosCacheFile).Encode(&latestVideos)
	}

	return latestVideos, nil
}

// UploadVideo uploads the video to YouTube and returns errors if the
// video or the thumbnail were not uploaded correctly.
func (yt *YouTubeAPI) UploadVideo(videoUploadData *VideoUploadData) (videoUploadError, thumbnailUploadError error) {
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       videoUploadData.Title,
			Description: videoUploadData.Description,
			CategoryId:  strconv.Itoa(categoryNameToId[videoUploadData.Category]),
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: videoUploadData.PrivacyStatus,
		},
	}

	// If publishAt was not specified, we should not schedule the video
	if len(videoUploadData.PublishAt) > 0 {
		upload.Status.PublishAt = videoUploadData.PublishAt
	}

	// The API returns a 400 Bad Request response if tags is an empty string.
	if len(videoUploadData.Tags) > 0 {
		upload.Snippet.Tags = videoUploadData.Tags
	}

	videoFile, videoUploadError := os.Open(videoUploadData.VideoPath)
	if videoUploadError != nil {
		return
	}
	defer videoFile.Close()

	videoCall := yt.service.Videos.Insert([]string{"snippet", "status"}, upload)
	videoResponse, videoUploadError := videoCall.Media(videoFile).Do()
	if videoUploadError != nil {
		return
	}

	thumbnailFile, thumbnailUploadError := os.Open(videoUploadData.ThumbnailPath)
	if thumbnailUploadError != nil {
		return
	}
	defer thumbnailFile.Close()

	thumbnailCall := yt.service.Thumbnails.Set(videoResponse.Id)
	_, thumbnailUploadError = thumbnailCall.Media(thumbnailFile).Do()

	return
}
