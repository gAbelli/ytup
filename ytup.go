package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	usr, _              = user.Current()
	DEFAULT_CONFIG_PATH = filepath.Join(usr.HomeDir, ".config", "ytup", "defaults.json")
)

var readCache = flag.Bool("r", false, "Read data from cache")

var categoryConversion = map[string]int{
	"Film & Animation":      1,
	"Autos & Vehicles":      2,
	"Music":                 10,
	"Pets & Animals":        15,
	"Sports":                17,
	"Short Movies":          18,
	"Travel & Events":       19,
	"Gaming":                20,
	"Videoblogging":         21,
	"People & Blogs":        22,
	"Comedy":                23,
	"Entertainment":         24,
	"News & Politics":       25,
	"Howto & Style":         26,
	"Education":             27,
	"Science & Technology":  28,
	"Nonprofits & Activism": 29,
	"Movies":                30,
	"Anime/Animation":       31,
	"Action/Adventure":      32,
	"Classics":              33,
	"Documentary":           35,
	"Drama":                 36,
	"Family":                37,
	"Foreign":               38,
	"Horror":                39,
	"Sci-Fi/Fantasy":        40,
	"Thriller":              41,
	"Shorts":                42,
	"Shows":                 43,
	"Trailers":              44,
}

type VideoUploadData struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Category      string   `json:"category"`
	PrivacyStatus string   `json:"privacy_status"`
	PublishAt     string   `json:"publish_at"`
}

type VideoDownloadData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	VideoId     string `json:"video_id"`
}

func Usage() {
	fmt.Printf("Usage: %s /path/to/video [/path/to/thumbnail]\n", os.Args[0])
	flag.PrintDefaults()
}

func parseArgs() (string, string) {
	flag.Usage = Usage
	flag.Parse()
	var videoPath, thumbnailPath string
	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(1)
	}
	_, err := os.Stat(flag.Arg(0))
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "The file %v does not exist\n", flag.Arg(0))
		os.Exit(1)
	} else if err != nil {
		panic(err)
	}
	videoPath = flag.Arg(0)
	if flag.NArg() == 2 {
		_, err := os.Stat(flag.Arg(1))
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "The file %v does not exist\n", flag.Arg(1))
			os.Exit(1)
		} else if err != nil {
			panic(err)
		}
		thumbnailPath = flag.Arg(1)
	}
	return videoPath, thumbnailPath
}

func main() {
	videoPath, thumbnailPath := parseArgs()

	var categories []string
	for category := range categoryConversion {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	privacyStatuses := []string{"Private", "Unlisted", "Public"}

	var defaultConfig VideoUploadData
	configFile, err := os.Open(DEFAULT_CONFIG_PATH)
	if err == nil {
		err = json.NewDecoder(configFile).Decode(&defaultConfig)
		if err != nil {
			panic(err)
		}
	}
	configFile.Close()

	defaultCategoryIndex := 0
	if _, ok := categoryConversion[defaultConfig.Category]; ok {
		for i, category := range categories {
			if category == defaultConfig.Category {
				defaultCategoryIndex = i
				break
			}
		}
	}
	defaultPrivacyStatusIndex := 0
	for i, status := range privacyStatuses {
		if status == defaultConfig.PrivacyStatus {
			defaultPrivacyStatusIndex = i
			break
		}
	}

	latestVideos, err := getLatestVideos(*readCache)
	if err != nil {
		panic(err)
	}

	app := tview.NewApplication()

	list := tview.NewList()
	digits := "1234567890"
	for i := 0; i < len(latestVideos); i++ {
		j := i
		list.AddItem(latestVideos[i].Title, "", rune(digits[i]), func() {
			defaultConfig.Title = latestVideos[j].Title
			defaultConfig.Description = latestVideos[j].Description
			tags, err := getVideoTags(latestVideos[j].VideoId)
			if err != nil {
				panic(err)
			}
			defaultConfig.Tags = tags
			app.Stop()
		})
	}
	list.AddItem("None", "", 'n', func() {
		app.Stop()
	})
	frame := tview.NewFrame(list).
		SetBorders(2, 2, 2, 2, 4, 4).
		AddText("Import data from a recent video", true, tview.AlignLeft, tcell.ColorRed)

	if err := app.SetRoot(frame, true).EnableMouse(true).SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			app.Stop()
			os.Exit(0)
		}
		return event
	}).Run(); err != nil {
		panic(err)
	}

	form := tview.NewForm().
		AddTextView("Video file", flag.Arg(0), 100, 1, true, false).
		AddTextView("Thumbnail file", flag.Arg(1), 100, 1, true, false).
		AddInputField("Title", defaultConfig.Title, 100, nil, nil).
		AddTextArea("Description", defaultConfig.Description, 100, 15, 0, nil).
		AddInputField("Tags (comma-separated)", strings.Join(defaultConfig.Tags, ","), 100, nil, nil).
		AddDropDown("Category", categories, defaultCategoryIndex, nil).
		AddDropDown("Privacy status", privacyStatuses, defaultPrivacyStatusIndex, nil).
		AddInputField("Publish at", time.Now().AddDate(0, 0, 1).Format(time.RFC3339), 100, nil, nil)

	titleItem, _ := form.GetFormItem(2).(*tview.InputField)
	descriptionItem, _ := form.GetFormItem(3).(*tview.TextArea)
	tagsItem, _ := form.GetFormItem(4).(*tview.InputField)
	categoryItem, _ := form.GetFormItem(5).(*tview.DropDown)
	privacyStatusItem, _ := form.GetFormItem(6).(*tview.DropDown)
	publishAtItem, _ := form.GetFormItem(7).(*tview.InputField)

	categoryItem.SetListStyles(tcell.StyleDefault.Background(tcell.ColorDarkSlateGrey), tcell.StyleDefault.Background(tcell.ColorDarkGreen))
	privacyStatusItem.SetListStyles(tcell.StyleDefault.Background(tcell.ColorDarkSlateGrey), tcell.StyleDefault.Background(tcell.ColorDarkGreen))

	form.
		AddButton("Upload", func() {
			var videoUploadData VideoUploadData

			videoUploadData.Title = titleItem.GetText()
			videoUploadData.Description = descriptionItem.GetText()
			videoUploadData.Tags = strings.Split(tagsItem.GetText(), ",")
			_, category := categoryItem.GetCurrentOption()
			videoUploadData.Category = category
			_, privacyStatus := privacyStatusItem.GetCurrentOption()
			videoUploadData.PrivacyStatus = strings.ToLower(privacyStatus)
			videoUploadData.PublishAt = publishAtItem.GetText()

			app.Stop()
			fmt.Println("Uploading...")
			fmt.Println(videoPath, thumbnailPath, videoUploadData)
			videoUploadError, thumbnailUploadError := uploadVideo(videoPath, thumbnailPath, videoUploadData)
			if videoUploadError != nil {
				fmt.Fprintf(os.Stderr, "There was an error in the video upload: %v\n", videoUploadError)
			} else {
				fmt.Println("Video uploaded successfully")
				if thumbnailUploadError == nil {
					fmt.Println("Thumbnail added successfully")
				}
			}
		})

	// Button styling does not work for some reason
	// uploadButton := form.GetButton(0)
	// uploadButton.SetStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))

	form.SetFieldBackgroundColor(tcell.GetColor("#606060")).SetBorder(true).SetTitle("Upload YouTube video").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
