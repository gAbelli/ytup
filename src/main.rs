use inquire;

mod client;
mod editor;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let youtube_client = client::YouTubeClient::new().await?;

    let latest_videos = youtube_client.get_last_n_videos(10).await?;
    if latest_videos.is_empty() {
        return Err(anyhow::anyhow!("No videos found"));
    }

    let selected_video = inquire::Select::new("Choose a video to import data from:", latest_videos)
        .with_vim_mode(true)
        .prompt()?;
    let selected_video = youtube_client.get_video_data(&selected_video.id).await?;

    let video_upload_request = editor::create_video_upload_request(selected_video);

    println!("Video upload request: {:?}", video_upload_request);

    Ok(())
}
