use anyhow::Result;

mod api;

#[tokio::main]
async fn main() -> Result<()> {
    let youtube_api = api::YouTubeApi::new().await?;
    let last_10_videos = youtube_api.get_last_videos(10).await?;

    println!("{:?}", last_10_videos);

    Ok(())
}
