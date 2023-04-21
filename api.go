// Sample Go code for user authorization

package main

import (
	"encoding/json"
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

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

func upload(service *youtube.Service, videoPath, thumbnailPath string, videoData VideoData) (videoUploadError, thumbnailUploadError error) {
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       videoData.Title,
			Description: videoData.Description,
			CategoryId:  strconv.Itoa(categoryConversion[videoData.Category]),
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: videoData.PrivacyStatus,
			// PublishAt: "2024-10-05T15:00:00.000Z",
		},
	}

	// The API returns a 400 Bad Request response if tags is an empty string.
	if len(videoData.Tags) > 0 {
		upload.Snippet.Tags = videoData.Tags
	}

	videoFile, videoUploadError := os.Open(videoPath)
	if videoUploadError != nil {
		return
	}
	defer videoFile.Close()

	thumbnailFile, thumbnailUploadError := os.Open(thumbnailPath)
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

func UploadVideo(videoPath, thumbnailPath string, videoData VideoData) (error, error) {
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
	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)
	service, err := youtube.New(client)

	handleError(err, "Error creating YouTube client")

	videoUploadError, thumbnailUploadError := upload(service, videoPath, thumbnailPath, videoData)
	return videoUploadError, thumbnailUploadError
}
