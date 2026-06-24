use crate::{Context, Error};
use poise::serenity_prelude as serenity;

#[poise::command(slash_command, prefix_command, guild_only)]
pub async fn join(ctx: Context<'_>) -> Result<(), Error> {
    let guild_id = ctx.guild_id().unwrap();
    let guild = ctx.serenity_context().cache.guild(guild_id).unwrap().clone();

    let channel_id = guild
        .voice_states
        .get(&ctx.author().id)
        .and_then(|vs| vs.channel_id);

    let channel_id = match channel_id {
        Some(c) => c,
        None => {
            ctx.say("You must be in a voice channel!").await?;
            return Ok(());
        }
    };

    let manager = songbird::get(ctx.serenity_context()).await.unwrap();
    match manager.join(guild_id, channel_id).await {
        Ok(_) => ctx.say("Joined your voice channel!").await?,
        Err(e) => ctx.say(format!("Error joining: {:?}", e)).await?,
    };

    Ok(())
}