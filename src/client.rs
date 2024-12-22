use anyhow::Context;
use chrono::{DateTime, Utc};
use google_youtube3::{
    api::{Video, VideoSnippet, VideoStatus},
    hyper::{self, client::HttpConnector},
    hyper_rustls::{self, HttpsConnector},
    oauth2, YouTube,
};
use serde::{Deserialize, Serialize};
use std::path::Path;

pub struct YouTubeClient {
    hub: YouTube<HttpsConnector<HttpConnector>>,
}

impl YouTubeClient {
    pub async fn new(
        client_secret_path: &std::path::Path,
        token_cache_path: &std::path::Path,
    ) -> anyhow::Result<YouTubeClient> {
        let secret = oauth2::read_application_secret(client_secret_path).await?;

        let auth = oauth2::InstalledFlowAuthenticator::builder(
            secret,
            oauth2::InstalledFlowReturnMethod::HTTPRedirect,
        )
        .persist_tokens_to_disk(token_cache_path)
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

        Ok(YouTubeClient { hub })
    }

    pub async fn get_last_n_videos(&self, n: u32) -> anyhow::Result<Vec<VideoSearchResponse>> {
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

    pub async fn get_video_data(&self, video_id: &str) -> anyhow::Result<VideoListResponse> {
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
            tags: snippet.tags.unwrap_or_default(),
            category: snippet.category_id.context("Could not retrieve category")?,
        };

        Ok(video_data)
    }

    pub async fn upload_video(
        &self,
        video_upload_request: VideoUploadRequest,
        video_path: &Path,
    ) -> anyhow::Result<String> {
        let mut video_request = Video::default();
        video_request.snippet = Some(VideoSnippet {
            title: Some(video_upload_request.title),
            description: Some(video_upload_request.description),
            tags: Some(video_upload_request.tags),
            category_id: Some(video_upload_request.category),
            ..Default::default()
        });
        video_request.status = Some(VideoStatus {
            privacy_status: Some(video_upload_request.privacy_status),
            publish_at: DateTime::parse_from_rfc3339(&video_upload_request.publish_at)
                .ok()
                .map(|dt| dt.with_timezone(&Utc)),
            ..Default::default()
        });

        let video_file = std::fs::File::open(video_path)?;
        let response = self
            .hub
            .videos()
            .insert(video_request)
            .upload(video_file, "application/octet-stream".parse().unwrap())
            .await?;

        let video_id = response.1.id.context("Could not retrieve video id")?;

        Ok(video_id)
    }

    pub async fn add_thumbnail(&self, video_id: &str, thumbnail_path: &Path) -> anyhow::Result<()> {
        let thumbnail_file = std::fs::File::open(thumbnail_path)?;
        self.hub
            .thumbnails()
            .set(video_id)
            .upload(thumbnail_file, "application/octet-stream".parse().unwrap())
            .await?;

        Ok(())
    }
}

#[derive(Debug)]
pub struct VideoSearchResponse {
    pub id: String,
    pub title: String,
}

impl std::fmt::Display for VideoSearchResponse {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.title)
    }
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
