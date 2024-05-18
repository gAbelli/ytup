use crate::client::{VideoListResponse, VideoUploadRequest};
use chrono::{Datelike, Days, Local, TimeZone};
use std::io::Read;

const DESCRIPTION: &str = r#"
# ytup - YouTube Uploader
# * Edit the video details above, then save and close the editor
# * Leave `publish_at` empty to avoid scheduling the video
# * Use the mapping below to set the `category` field
# 
#     "Film & Animation":      1
#     "Autos & Vehicles":      2
#     "Music":                 10
#     "Pets & Animals":        15
#     "Sports":                17
#     "Short Movies":          18
#     "Travel & Events":       19
#     "Gaming":                20
#     "Videoblogging":         21
#     "People & Blogs":        22
#     "Comedy":                23
#     "Entertainment":         24
#     "News & Politics":       25
#     "Howto & Style":         26
#     "Education":             27
#     "Science & Technology":  28
#     "Nonprofits & Activism": 29
#     "Movies":                30
#     "Anime/Animation":       31
#     "Action/Adventure":      32
#     "Classics":              33
#     "Documentary":           35
#     "Drama":                 36
#     "Family":                37
#     "Foreign":               38
#     "Horror":                39
#     "Sci-Fi/Fantasy":        40
#     "Thriller":              41
#     "Shorts":                42
#     "Shows":                 43
#     "Trailers":              44
"#;

pub fn create_video_upload_request(
    source_video: VideoListResponse,
) -> anyhow::Result<VideoUploadRequest> {
    let video_upload_request = VideoUploadRequest {
        title: source_video.title,
        description: source_video.description,
        tags: source_video.tags,
        category: source_video.category,
        privacy_status: "private".to_owned(),
        publish_at: tomorrow(),
    };
    let mut video_upload_request = serde_yaml::to_string(&video_upload_request)?;
    video_upload_request.push_str(DESCRIPTION);

    let yaml_file_path = std::env::temp_dir().join("video_upload_data.yaml");
    std::fs::write(&yaml_file_path, video_upload_request)?;

    let content = edit(&yaml_file_path)?;

    let video_upload_request: VideoUploadRequest = serde_yaml::from_str(&content)?;

    Ok(video_upload_request)
}

fn edit(file_path: &std::path::Path) -> anyhow::Result<String> {
    let editor = std::env::var("EDITOR").unwrap_or("vim".to_owned());
    std::process::Command::new(editor)
        .arg(&file_path)
        .status()?;
    let mut content = String::new();
    std::fs::File::open(file_path)?.read_to_string(&mut content)?;

    Ok(content)
}

fn tomorrow() -> String {
    let now = Local::now();
    let tomorrow = Local
        .with_ymd_and_hms(now.year(), now.month(), now.day(), 0, 0, 0)
        .unwrap()
        .checked_add_days(Days::new(1))
        .unwrap()
        .to_rfc3339();

    tomorrow
}
