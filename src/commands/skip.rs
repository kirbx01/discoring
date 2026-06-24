use crate::{Context, Error};

#[poise::command(slash_command, prefix_command, guild_only)]
pub async fn skip(ctx: Context<'_>) -> Result<(), Error> {
    let guild_id = ctx.guild_id().unwrap();
    let manager = songbird::get(ctx.serenity_context()).await.unwrap();

    if let Some(handler_lock) = manager.get(guild_id) {
        let handler = handler_lock.lock().await;
        let queue = handler.queue();
        if queue.is_empty() {
            ctx.say("Nothing is playing!").await?;
        } else {
            queue.skip()?;
            ctx.say(format!("Skipped! {} track(s) remaining.", queue.len().saturating_sub(1))).await?;
        }
    } else {
        ctx.say("I'm not in a voice channel!").await?;
    }

    Ok(())
}