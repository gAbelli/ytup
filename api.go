package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".config", "ytup")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-api.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getService() (*youtube.Service, error) {
	ctx := context.Background()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile(homeDir + "/.config/ytup/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.config/ytup/youtube-api.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubeForceSslScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)
	service, err := youtube.New(client)
	return service, err
}

func GetVideoTags(id string) ([]string, error) {
	service, err := getService()
	if err != nil {
		log.Fatalf(err.Error())
	}

	listCall := service.Videos.List([]string{"snippet"})
	listResponse, err := listCall.Id(id).Do()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if len(listResponse.Items) < 1 {
		return nil, errors.New("Video not found")
	}

	return listResponse.Items[0].Snippet.Tags, nil
}

func GetLatestVideos(readCache bool) ([]VideoDownloadData, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	cacheFilePath := filepath.Join(usr.HomeDir, ".config", "ytup", "videos_cache.json")
	_, err = os.Stat(cacheFilePath)
	if os.IsNotExist(err) {
		os.Create(cacheFilePath)
	} else if err != nil {
		panic(err)
	}

	videosCacheFile, err := os.OpenFile(cacheFilePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	defer videosCacheFile.Close()

	var latestVideos []VideoDownloadData
	if readCache {
		json.NewDecoder(videosCacheFile).Decode(&latestVideos)
	} else {
		service, err := getService()
		if err != nil {
			log.Fatalf(err.Error())
		}

		listCall := service.Search.List([]string{"snippet"})
		listResponse, err := listCall.ForMine(true).MaxResults(10).Order("date").Type("video").Do()
		if err != nil {
			return nil, err
		}

		for _, video := range listResponse.Items {
			latestVideos = append(latestVideos, VideoDownloadData{
				Title:       video.Snippet.Title,
				Description: video.Snippet.Description,
				VideoId:     video.Id.VideoId,
			})
		}
		json.NewEncoder(videosCacheFile).Encode(&latestVideos)
	}

	return latestVideos, nil
}

func UploadVideo(videoUploadData *VideoUploadData) (videoUploadError, thumbnailUploadError error) {
	service, err := getService()
	if err != nil {
		log.Fatalf(err.Error())
	}

	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       videoUploadData.Title,
			Description: videoUploadData.Description,
			CategoryId:  strconv.Itoa(categoryConversion[videoUploadData.Category]),
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: videoUploadData.PrivacyStatus,
			PublishAt:     videoUploadData.PublishAt,
		},
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

	thumbnailFile, thumbnailUploadError := os.Open(videoUploadData.ThumbnailPath)
	if thumbnailUploadError != nil && !os.IsNotExist(thumbnailUploadError) {
		return
	}
	defer thumbnailFile.Close()

	videoCall := service.Videos.Insert([]string{"snippet", "status"}, upload)
	videoResponse, videoUploadError := videoCall.Media(videoFile).Do()
	if videoUploadError != nil {
		return
	}

	if thumbnailUploadError == nil {
		thumbnailCall := service.Thumbnails.Set(videoResponse.Id)
		_, thumbnailUploadError = thumbnailCall.Media(thumbnailFile).Do()
	}

	return
}
