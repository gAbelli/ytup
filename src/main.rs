use anyhow::Result;
use inquire;

mod api;

#[tokio::main]
async fn main() -> Result<()> {
    // let youtube_api = api::YouTubeApi::new().await?;
    // let last_10_videos = youtube_api.get_last_videos(10).await?;
    //
    // let mut options: Vec<_> = last_10_videos
    //     .iter()
    //     .enumerate()
    //     .map(|(i, video)| inquire::list_option::ListOption::new(i, video.title.as_str()))
    //     .collect();
    //
    // options.push(inquire::list_option::ListOption::new(options.len(), "None"));

    let options = vec![
        inquire::list_option::ListOption::new(0, "Nazionali 2024 squadre 11"),
        inquire::list_option::ListOption::new(1, "Nazionali 2024 individuali 3"),
        inquire::list_option::ListOption::new(2, "Nazionali 2024 individuali 2"),
    ];

    let selected = inquire::Select::new("Choose a video to import data from:", options)
        .with_vim_mode(true)
        .prompt()?;

    println!("Selected video: {}", selected);

    Ok(())
}
