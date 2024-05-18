use anyhow::Result;
use chrono::{Datelike, Days, Local, TimeZone};
use inquire;
use std::io::Read;

mod api;

pub fn edit(file_path: &str) -> Result<String> {
    let editor = std::env::var("EDITOR")?;
    std::process::Command::new(editor)
        .arg(&file_path)
        .status()?;
    let mut content = String::new();
    std::fs::File::open(file_path)?.read_to_string(&mut content)?;

    Ok(content)
}

#[tokio::main]
async fn main() -> Result<()> {
    let youtube_api = api::YouTubeApi::new().await?;
    // let last_10_videos = youtube_api.get_last_videos(10).await?;

    // if last_10_videos.is_empty() {
    //     return Err(anyhow::anyhow!("No videos found"));
    // }

    impl std::fmt::Display for api::VideoSearchResponse {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "{}", self.title)
        }
    }

    // let selected = inquire::Select::new("Choose a video to import data from:", last_10_videos)
    //     .with_vim_mode(true)
    //     .prompt()?;

    let video_id = "nufv6UzW8U0";

    let video_data = youtube_api.get_video_data(video_id).await?;

    let now = Local::now();
    let tomorrow = Local
        .with_ymd_and_hms(now.year(), now.month(), now.day(), 0, 0, 0)
        .unwrap()
        .checked_add_days(Days::new(1))
        .unwrap()
        .to_rfc3339();

    let video_upload_request = api::VideoUploadRequest {
        title: video_data.title,
        description: video_data.description,
        tags: video_data.tags,
        category: video_data.category,
        privacy_status: "private".to_owned(),
        publish_at: tomorrow,
    };

    let yaml = serde_yaml::to_string(&video_upload_request)?;

    let yaml_file_path = std::env::temp_dir().join("video_upload_data.yaml");
    println!("Writing video data to {}", yaml_file_path.display());
    std::fs::write(&yaml_file_path, yaml)?;

    let yaml_file_path = yaml_file_path.to_str().unwrap();
    let content = edit(yaml_file_path)?;

    let video_upload_request: api::VideoUploadRequest = serde_yaml::from_str(&content)?;

    println!("Video upload request: {:?}", video_upload_request);

    Ok(())
}
