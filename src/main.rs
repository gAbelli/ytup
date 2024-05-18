use clap::Parser;

mod client;
mod editor;

#[derive(Debug, Parser)]
struct Args {
    video_path: std::path::PathBuf,
    thumbnail_path: std::path::PathBuf,
    #[arg(long)]
    client_secret_path: Option<std::path::PathBuf>,
    #[arg(long)]
    token_cache_path: Option<std::path::PathBuf>,
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    let args = Args::parse();
    let home_dir = dirs::home_dir().unwrap();

    let client_secret_path = args
        .client_secret_path
        .unwrap_or(home_dir.join(".config/ytup/client_secret.json"));
    let token_cache_path = args
        .token_cache_path
        .unwrap_or(home_dir.join(".local/share/ytup/token_cache.json"));

    let youtube_client = client::YouTubeClient::new(&client_secret_path, &token_cache_path).await?;

    let latest_videos = youtube_client.get_last_n_videos(10).await?;
    if latest_videos.is_empty() {
        return Err(anyhow::anyhow!("No videos found"));
    }

    let source_video = inquire::Select::new("Choose a video to import data from:", latest_videos)
        .with_vim_mode(true)
        .prompt()?;
    let source_video = youtube_client.get_video_data(&source_video.id).await?;

    let video_upload_request = editor::create_video_upload_request(source_video)?;

    println!("Uploading video...");
    let video_id = youtube_client
        .uplaod_video(video_upload_request, &args.video_path)
        .await?;
    println!("Video uploaded");

    println!("Adding thumbnail...");
    youtube_client
        .add_thumbnail(&video_id, &args.thumbnail_path)
        .await?;
    println!("Thumbnail added");

    Ok(())
}
