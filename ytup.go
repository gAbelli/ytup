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

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	usr, _              = user.Current()
	DEFAULT_CONFIG_PATH = filepath.Join(usr.HomeDir, ".config", "ytup", "defaults.json")
)

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

type VideoData struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Category      string   `json:"category"`
	PrivacyStatus string   `json:"privacy_status"`
	PublishAt     string   `json:"publish_at"`
}

func Usage() {
	fmt.Printf("Usage: %s [OPTIONS] /path/to/video [/path/to/thumbnail]\n", os.Args[0])
}

func main() {
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

	var categories []string
	for category := range categoryConversion {
		categories = append(categories, category)
	}
	sort.Strings(categories)

	privacyStatuses := []string{"private", "unlisted", "public"}

	var defaultConfig VideoData
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

	app := tview.NewApplication()
	form := tview.NewForm().
		AddTextView("Video path", flag.Arg(0), 100, 1, true, false).
		AddTextView("Thumbnail path", flag.Arg(1), 100, 1, true, false).
		AddInputField("Title", defaultConfig.Title, 100, nil, nil).
		AddTextArea("Description", defaultConfig.Description, 100, 0, 0, nil).
		AddInputField("Tags (comma-separated)", strings.Join(defaultConfig.Tags, ","), 100, nil, nil).
		AddDropDown("Category", categories, defaultCategoryIndex, nil).
		AddDropDown("Privacy status", privacyStatuses, defaultPrivacyStatusIndex, nil)
	form.
		AddButton("Upload", func() {
			var videoData VideoData

			fmt.Println(form.GetFormItemCount())
			titleItem, _ := form.GetFormItem(2).(*tview.InputField)
			videoData.Title = titleItem.GetText()

			descriptionItem, _ := form.GetFormItem(3).(*tview.TextArea)
			videoData.Description = descriptionItem.GetText()

			tagsItem, _ := form.GetFormItem(4).(*tview.InputField)
			videoData.Tags = strings.Split(tagsItem.GetText(), ",")

			categoryItem, _ := form.GetFormItem(5).(*tview.DropDown)
			_, category := categoryItem.GetCurrentOption()
			videoData.Category = category

			privacyStatusItem, _ := form.GetFormItem(6).(*tview.DropDown)
			_, privacyStatus := privacyStatusItem.GetCurrentOption()
			videoData.PrivacyStatus = privacyStatus

			app.Stop()
			fmt.Println("Uploading...")
			videoUploadError, thumbnailUploadError := UploadVideo(videoPath, thumbnailPath, videoData)
			if videoUploadError != nil {
				fmt.Fprintf(os.Stderr, "There was an error in the video upload: %v\n", videoUploadError)
			} else {
				fmt.Println("Video uploaded succesfully")
			}
			if thumbnailUploadError == nil {
				fmt.Println("Thumbnail added succesfully")
			}
		})
	form.SetFieldBackgroundColor(tcell.GetColor("darkcyan")).SetBorder(true).SetTitle("Upload YouTube video").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
