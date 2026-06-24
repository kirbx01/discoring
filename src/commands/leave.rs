use crate::{Context, Error};

#[poise::command(slash_command, prefix_command, guild_only)]
pub async fn leave(ctx: Context<'_>) -> Result<(), Error> {
    let guild_id = ctx.guild_id().unwrap();
    let manager = songbird::get(ctx.serenity_context()).await.unwrap();

    if manager.get(guild_id).is_some() {
        if let Some(handler_lock) = manager.get(guild_id) {
            handler_lock.lock().await.queue().stop();
        }
        manager.remove(guild_id).await?;
        ctx.say("Left and cleared the queue!").await?;
    } else {
        ctx.say("I'm not in a voice channel!").await?;
    }

    Ok(())
}