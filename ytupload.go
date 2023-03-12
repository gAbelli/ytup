package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type VideoInfo struct {
	VideoPath, ThumbnailPath, Title, Description string
	Tags                                         []string
}

func getDefaultsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	defaultsPath := homeDir + "/.config/ytup/defaults.txt"
	_, error := os.Stat(defaultsPath)
	if os.IsNotExist(error) {
		err := os.MkdirAll(homeDir + "/.config/ytup/", os.ModePerm)
		if err != nil {
			return "", err
		}
		f, err := os.Create(defaultsPath)
		if err != nil {
			return "", err
		}
		defer f.Close()
		_, err = f.WriteString("# title\n# description\n# tags")
		if err != nil {
			return "", err
		}
	}
	return defaultsPath, nil
}

func parseDefaults(filePath string) []string {
	const (
		readingTitle       = 0
		readingDescription = 1
		readingTags        = 2
	)
	strs := []string{"", "", ""}
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	state := readingTitle
	for scanner.Scan() {
		s := scanner.Text()
		switch s {
		case "# title":
			state = readingTitle
		case "# description":
			state = readingDescription
		case "# tags":
			state = readingTags
		default:
			if len(strs[state]) > 0 {
				strs[state] += "\n"
			}
			strs[state] += s
		}
		if err != nil {
			break
		}
	}
	return strs
}

var editDefaults = flag.Bool("e", false, "edit defaults")

func main() {
	flag.Parse()

	defaultsPath, err := getDefaultsPath()
	if err != nil {
		panic(err)
	}

	if *editDefaults {
		defaultEditor, ok := os.LookupEnv("EDITOR")
		if !ok {
			defaultEditor = "vim"
		}
		c := exec.Command(defaultEditor, defaultsPath)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		if err != nil {
			panic(err)
		}
		return
	}

	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %v /path/to/video /path/to/thumbnail\n", os.Args[0])
		os.Exit(1)
	}

	app := tview.NewApplication()

	strs := parseDefaults(defaultsPath)
	shouldUpload := false

	form := tview.NewForm().
		AddInputField("Title", strs[0], 100, nil, func(text string) {
			strs[0] = text
		}).
		AddTextArea("Description", strs[1], 100, 20, 0, func(text string) {
			strs[1] = text
		}).
		AddTextArea("Tags", strs[2], 100, 1, 0, func(text string) {
			strs[2] = text
		}).
		AddButton("Upload", func() {
			shouldUpload = true
			app.Stop()
		})
	form.SetFieldBackgroundColor(tcell.GetColor("#263238"))
	form.SetFieldTextColor(tcell.GetColor("#ffffff"))
	form.SetButtonBackgroundColor(tcell.GetColor("#2E3C43"))

	form.SetBorder(true).SetTitle("Upload a video to YouTube").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
	if !shouldUpload {
		return
	}

	videoInfo := VideoInfo{
		VideoPath:     flag.Arg(0),
		ThumbnailPath: flag.Arg(1),
		Title:         strs[0],
		Description:   strs[1],
		Tags:          strings.Split(strs[2], ","),
	}
	UploadVideo(videoInfo)
}
