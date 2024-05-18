use anyhow::Result;
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

    pub async fn get_last_videos(&self, n: u32) -> Result<Vec<(String, String)>> {
        let last_ten_videos = self
            .hub
            .search()
            .list(&vec!["snippet".into(), "id".into()])
            .add_type("video")
            .for_mine(true)
            .max_results(n)
            .order("date")
            .doit()
            .await?;

        let items = last_ten_videos.1.items.unwrap();

        let ids_and_titles: Vec<_> = items
            .iter()
            .map(|item| {
                (
                    item.id
                        .as_ref()
                        .unwrap()
                        .video_id
                        .as_ref()
                        .unwrap()
                        .to_owned(),
                    item.snippet
                        .as_ref()
                        .unwrap()
                        .title
                        .as_ref()
                        .unwrap()
                        .to_owned(),
                )
            })
            .collect();

        Ok(ids_and_titles)
    }
}
