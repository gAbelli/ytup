package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/rivo/tview"
)

func parseDefaults(dir string) []string {
	const (
		readingTitle       = 0
		readingDescription = 1
		readingTags        = 2
	)
	strs := []string{"", "", ""}
	f, err := os.Open(dir)
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

func main() {
	app := tview.NewApplication()

	strs := parseDefaults("./defaults.txt")

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
		AddButton("Save", func() {
			for i := 0; i < 3; i++ {
				fmt.Fprintf(os.Stderr, "%v\n", strs[i])
			}
		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Enter some data").SetTitleAlign(tview.AlignLeft)
	if err := app.SetRoot(form, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
