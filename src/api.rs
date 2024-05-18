use anyhow::{Context, Result};
use google_youtube3::{
    hyper, hyper::client::HttpConnector, hyper_rustls, hyper_rustls::HttpsConnector, oauth2,
    YouTube,
};

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

    pub async fn get_last_videos(&self, n: u32) -> Result<Vec<VideoSearchResult>> {
        Ok(self
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

                Ok::<VideoSearchResult, anyhow::Error>(VideoSearchResult {
                    id: video_id,
                    title: title,
                })
            })
            .collect())
    }
}

pub struct VideoSearchResult {
    pub id: String,
    pub title: String,
}
