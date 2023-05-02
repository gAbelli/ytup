package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	log.SetFlags(0)
	videoPath, thumbnailPath := ParseArgs()
	formData, err := GetDefaults()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	yt, err := NewYouTubeAPI()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	latestVideos, err := yt.GetLatestVideos(*readCache)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	app := tview.NewApplication()
	list := tview.NewList()
	digits := "1234567890"
	for i := 0; i < len(latestVideos); i++ {
		j := i
		list.AddItem(latestVideos[i].Title, "", rune(digits[i]), func() {
			formData.Title = latestVideos[j].Title
			formData.Description = latestVideos[j].Description
			extraVideoData, err := yt.GetExtraVideoData(latestVideos[j].VideoId)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			formData.Tags = extraVideoData.Tags
			for i, category := range categories {
				if category == extraVideoData.Category {
					formData.CategoryIndex = i
					break
				}
			}
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
		log.Fatalf("Error: %v", err)
	}

	videoFileName := path.Base(flag.Arg(0))
	thumbnailFileName := ""
	if flag.NArg() == 2 {
		thumbnailFileName = path.Base(flag.Arg(1))
	}

	form := tview.NewForm().
		AddTextView("Video file", videoFileName, 100, 1, true, false).
		AddTextView("Thumbnail file", thumbnailFileName, 100, 1, true, false).
		AddInputField("Title", formData.Title, 100, nil, nil).
		AddTextArea("Description", formData.Description, 100, 15, 0, nil).
		AddInputField("Tags (comma-separated)", strings.Join(formData.Tags, ","), 100, nil, nil).
		AddDropDown("Category", categories, formData.CategoryIndex, nil).
		AddDropDown("Privacy status", upperCasePrivacyStatuses, formData.PrivacyStatusIndex, nil).
		AddInputField("Publish at", formData.PublishAt, 100, nil, nil)

	titleItem, _ := form.GetFormItem(2).(*tview.InputField)
	descriptionItem, _ := form.GetFormItem(3).(*tview.TextArea)
	tagsItem, _ := form.GetFormItem(4).(*tview.InputField)
	categoryItem, _ := form.GetFormItem(5).(*tview.DropDown)
	privacyStatusItem, _ := form.GetFormItem(6).(*tview.DropDown)
	publishAtItem, _ := form.GetFormItem(7).(*tview.InputField)

	categoryItem.SetListStyles(tcell.StyleDefault.Background(tcell.GetColor("#303030")), tcell.StyleDefault.Background(tcell.ColorDarkGreen))
	privacyStatusItem.SetListStyles(tcell.StyleDefault.Background(tcell.GetColor("#303030")), tcell.StyleDefault.Background(tcell.ColorDarkGreen))

	form.
		AddButton("Upload", func() {
			var videoUploadData VideoUploadData

			videoUploadData.VideoPath = videoPath
			videoUploadData.ThumbnailPath = thumbnailPath
			videoUploadData.Title = titleItem.GetText()
			videoUploadData.Description = descriptionItem.GetText()
			videoUploadData.Tags = strings.Split(tagsItem.GetText(), ",")
			_, category := categoryItem.GetCurrentOption()
			videoUploadData.Category = category
			// Watch out for uppercase first letter
			privacyStatusIndex, _ := privacyStatusItem.GetCurrentOption()
			videoUploadData.PrivacyStatus = privacyStatuses[privacyStatusIndex]
			videoUploadData.PublishAt = publishAtItem.GetText()

			app.Stop()
			fmt.Println("Uploading...")
			videoUploadError, thumbnailUploadError := yt.UploadVideo(&videoUploadData)
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

	form.SetFieldBackgroundColor(tcell.ColorDarkSlateGrey).SetBorder(true).SetTitle("Upload YouTube video").SetTitleAlign(tview.AlignLeft)
	// form.SetFieldBackgroundColor(tcell.GetColor("#606060")).SetBorder(true).SetTitle("Upload YouTube video").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
