use std::env;
use poise::serenity_prelude as serenity;
use songbird::SerenityInit;

mod commands;

pub struct Data {}
pub type Error = Box<dyn std::error::Error + Send + Sync>;
pub type Context<'a> = poise::Context<'a, Data, Error>;

#[tokio::main]
async fn main() {
    dotenv::dotenv().ok();
    let token = env::var("DISCORD_TOKEN").expect("Missing DISCORD_TOKEN");

    let intents = serenity::GatewayIntents::non_privileged()
        | serenity::GatewayIntents::MESSAGE_CONTENT
        | serenity::GatewayIntents::GUILD_VOICE_STATES;

    let framework = poise::Framework::builder()
        .options(poise::FrameworkOptions {
            commands: vec![
                commands::join::join(),
                commands::play::play(),
                commands::skip::skip(),
                commands::leave::leave(),
            ],
            prefix_options: poise::PrefixFrameworkOptions {
                prefix: Some("!".into()),
                ..Default::default()
            },
            ..Default::default()
        })
        .setup(|ctx, _ready, framework| {
            Box::pin(async move {
                poise::builtins::register_globally(ctx, &framework.options().commands).await?;
                println!("✅ Bot is ready!");
                Ok(Data {})
            })
        })
        .build();

    serenity::ClientBuilder::new(token, intents)
        .framework(framework)
        .register_songbird()
        .await
        .unwrap()
        .start()
        .await
        .unwrap();
}
