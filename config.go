package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// Usage prints usage instructions to the user.
func Usage() {
	fmt.Printf("Usage: %s /path/to/video [/path/to/thumbnail]\n", os.Args[0])
	flag.PrintDefaults()
}

// ParseArgs parses the command line arguments.
func ParseArgs() (string, string) {
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

// GetDefaults reads the default configuration from a json file
// and returns the corresponding form data. If the file does not
// exist, the function still returns a default configuration
// and does not return any error.
func GetDefaults() (*FormData, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	configFilePath := filepath.Join(usr.HomeDir, ".config", "ytup", "defaults.json")

	var defaultConfig ConfigData
	configFile, err := os.Open(configFilePath)
	defer configFile.Close()
	if err == nil {
		err = json.NewDecoder(configFile).Decode(&defaultConfig)
		if err != nil {
			return nil, err
		}
	}

	// Go does not have a built-in function to search for an
	// element in a slice
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

	if len(defaultConfig.PublishTime) == 0 {
		defaultConfig.PublishTime = "0000"
	}
	publishTime, err := time.ParseDuration(defaultConfig.PublishTime[:2] + "h" + defaultConfig.PublishTime[2:4] + "m")
	if err != nil {
		return nil, err
	}
	// Hack to get tomorow's timestamp at midnight in the correct timezone
	year, month, day := time.Now().Date()
	tomorrow := time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location()).AddDate(0, 0, 1)

	publishAt := tomorrow.Add(publishTime).Format(time.RFC3339)

	formData := FormData{
		Title:              defaultConfig.Title,
		Description:        defaultConfig.Description,
		Tags:               defaultConfig.Tags,
		CategoryIndex:      defaultCategoryIndex,
		PrivacyStatusIndex: defaultPrivacyStatusIndex,
		PublishAt:          publishAt,
	}

	return &formData, nil
}
