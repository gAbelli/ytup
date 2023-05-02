package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Usage prints usage instructions to the user.
func Usage() {
	fmt.Printf("Usage: %s /path/to/video [/path/to/thumbnail]\n", os.Args[0])
	flag.PrintDefaults()
}

// checkIsFile checks if path is a file and returns an
// error if it isn't
func checkIsFile(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("The file %v does not exist", path)
	} else if info.IsDir() {
		return fmt.Errorf("%v is a directory", path)
	} else if err != nil {
		return err
	}
	return nil
}

// ParseArgs parses the command line arguments.
func ParseArgs() (string, string) {
	flag.Usage = Usage
	flag.Parse()
	var videoPath, thumbnailPath string

	// We must have either 1 or 2 arguments
	if flag.NArg() == 0 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(1)
	}

	// The video file is mandatory
	err := checkIsFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	videoPath = flag.Arg(0)

	// The thumbnail file, instead, is not mandatory
	if flag.NArg() == 2 {
		err = checkIsFile(flag.Arg(1))
		if err != nil {
			log.Fatalf("Error: %v", err)
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configFilePath := filepath.Join(homeDir, ".config", "ytup", "defaults.json")

	var defaultConfig ConfigData
	configFile, err := os.Open(configFilePath)
	if err == nil {
		defer configFile.Close()
		err = json.NewDecoder(configFile).Decode(&defaultConfig)
		if err != nil {
			return nil, err
		}
	}

	// Go does not have a built-in function to search for an
	// element in a slice
	defaultCategoryIndex := 0
	if _, ok := categoryNameToId[defaultConfig.Category]; ok {
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

	if len(defaultConfig.PublishTime) != 4 {
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
