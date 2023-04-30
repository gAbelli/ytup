package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func Usage() {
	fmt.Printf("Usage: %s /path/to/video [/path/to/thumbnail]\n", os.Args[0])
	flag.PrintDefaults()
}

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
			panic(err)
		}
	}

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

	formData := FormData{
		Title:              defaultConfig.Title,
		Description:        defaultConfig.Description,
		Tags:               defaultConfig.Tags,
		CategoryIndex:      defaultCategoryIndex,
		PrivacyStatusIndex: defaultPrivacyStatusIndex,
	}

	return &formData, nil
}
