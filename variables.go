package main

import (
	"flag"
)

var readCache = flag.Bool("r", false, "Read data from cache")

// A map that converts category names to their id on YouTube
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

// The slice of category strings sorted alphabetically
var categories = []string{
	"Action/Adventure",
	"Anime/Animation",
	"Autos & Vehicles",
	"Classics",
	"Comedy",
	"Documentary",
	"Drama",
	"Education",
	"Entertainment",
	"Family",
	"Film & Animation",
	"Foreign",
	"Gaming",
	"Horror",
	"Howto & Style",
	"Movies",
	"Music",
	"News & Politics",
	"Nonprofits & Activism",
	"People & Blogs",
	"Pets & Animals",
	"Science & Technology",
	"Sci-Fi/Fantasy",
	"Short Movies",
	"Shorts",
	"Shows",
	"Sports",
	"Thriller",
	"Trailers",
	"Travel & Events",
	"Videoblogging",
}

// We show the upper case version in the UI, but we use the lower case
// version internally
var (
	privacyStatuses          = []string{"private", "unlisted", "public"}
	upperCasePrivacyStatuses = []string{"Private", "Unlisted", "Public"}
)
