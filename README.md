# ytup - YouTube Uploader

A simple CLI to quickly upload videos to YouTube, written in Rust.

## Features

- Upload a video and its thumbnail from the command line.
- Import the title, description, tags and category from a recently uploaded video to avoid repetitive work, then adjust them from your editor.

## Usage

Place your `client_secret.json` file in `~/.config/ytup/`.

```
ytup [OPTIONS] <VIDEO_PATH> <THUMBNAIL_PATH>

Arguments:
  <VIDEO_PATH>
  [THUMBNAIL_PATH]

Options:
      --client-secret-path <CLIENT_SECRET_PATH>
      --token-cache-path <TOKEN_CACHE_PATH>
  -h, --help
```

## Quota usage

- Daily available quota: 10,000
- Video upload quota: 1,600
- Video search quota: 100

## Subscribe!

While you are here, subscribe to my YouTube channel [mateMATTIci](https://www.youtube.com/@mateMATTIci)!
