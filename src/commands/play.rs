use crate::{Context, Error};
use songbird::input::YoutubeDl;

#[poise::command(slash_command, prefix_command, guild_only)]
pub async fn play(
    ctx: Context<'_>,
    #[description = "YouTube URL or search term"] url: String,
) -> Result<(), Error> {
    let guild_id = ctx.guild_id().unwrap();
    let manager = songbird::get(ctx.serenity_context()).await.unwrap();

    if let Some(handler_lock) = manager.get(guild_id) {
        let mut handler = handler_lock.lock().await;
        let client = reqwest::Client::new();
        let source = YoutubeDl::new(client, url);
        handler.enqueue_input(source.into()).await;

        let len = handler.queue().len();
        if len > 1 {
            ctx.say(format!("Added to queue! Position: {}", len)).await?;
        } else {
            ctx.say("Now playing!").await?;
        }
    } else {
        ctx.say("I'm not in a voice channel! Use `/join` first.").await?;
    }

    Ok(())
}