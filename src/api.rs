use anyhow::{Context, Result};
use google_youtube3::{
    hyper, hyper::client::HttpConnector, hyper_rustls, hyper_rustls::HttpsConnector, oauth2,
    YouTube,
};
use serde::{Deserialize, Serialize};

pub struct YouTubeApi {
    hub: YouTube<HttpsConnector<HttpConnector>>,
}

impl YouTubeApi {
    pub async fn new() -> Result<YouTubeApi> {
        let secret =
            oauth2::read_application_secret("/Users/giorgio/.config/ytup/client_secret.json")
                .await?;

        let auth = oauth2::InstalledFlowAuthenticator::builder(
            secret,
            oauth2::InstalledFlowReturnMethod::HTTPRedirect,
        )
        .persist_tokens_to_disk("/Users/giorgio/.local/share/ytup/token_cache.json")
        .build()
        .await?;

        let hub = YouTube::new(
            hyper::Client::builder().build(
                hyper_rustls::HttpsConnectorBuilder::new()
                    .with_native_roots()
                    .https_or_http()
                    .enable_http1()
                    .build(),
            ),
            auth,
        );

        Ok(YouTubeApi { hub })
    }

    pub async fn get_last_videos(&self, n: u32) -> Result<Vec<VideoSearchResponse>> {
        let videos: Vec<_> = self
            .hub
            .search()
            .list(&vec!["snippet".into(), "id".into()])
            .add_type("video")
            .for_mine(true)
            .max_results(n)
            .order("date")
            .doit()
            .await?
            .1
            .items
            .context("Could not retrieve search results from the API")?
            .into_iter()
            .flat_map(|item| {
                let resource_id = item.id.context("Could not retrieve resource id")?;
                let video_id = resource_id
                    .video_id
                    .context("Could not retrieve video id")?;

                let snippet = item
                    .snippet
                    .context(format!("Could not retrieve snippet for video {}", video_id))?;
                let title = snippet
                    .title
                    .context(format!("Could not retrieve title for video {}", video_id))?;

                Ok::<VideoSearchResponse, anyhow::Error>(VideoSearchResponse {
                    id: video_id,
                    title,
                })
            })
            .collect();

        Ok(videos)
    }

    pub async fn get_video_data(&self, video_id: &str) -> Result<VideoListResponse> {
        let video = self
            .hub
            .videos()
            .list(&vec!["snippet".into()])
            .add_id(video_id)
            .doit()
            .await?
            .1
            .items
            .context("Could not retrieve video data from the API")?
            .into_iter()
            .next()
            .context("Could not retrieve video data from the API")?;

        let snippet = video.snippet.context("Could not retrieve snippet")?;

        let video_data = VideoListResponse {
            id: video_id.to_owned(),
            title: snippet.title.context("Could not retrieve title")?,
            description: snippet
                .description
                .context("Could not retrieve description")?,
            tags: snippet.tags.context("Could not retrieve tags")?,
            category: snippet.category_id.context("Could not retrieve category")?,
        };

        Ok(video_data)
    }
}

#[derive(Debug)]
pub struct VideoSearchResponse {
    pub id: String,
    pub title: String,
}

#[derive(Debug)]
pub struct VideoListResponse {
    pub id: String,
    pub title: String,
    pub description: String,
    pub tags: Vec<String>,
    pub category: String,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct VideoUploadRequest {
    pub title: String,
    pub description: String,
    pub tags: Vec<String>,
    pub category: String,
    pub privacy_status: String,
    pub publish_at: String,
}
