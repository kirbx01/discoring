use serenity::framework::standard::{macros::command, Args, CommandResult};
use serenity::model::prelude::*;
use serenity::prelude::*;

#[command]
#[only_in(guilds)]
async fn skip(ctx: &Context, msg: &Message, _args: Args) -> CommandResult {
    let guild_id = msg.guild_id.unwrap();
    let manager = songbird::get(ctx).await.unwrap();

    if let Some(handler_lock) = manager.get(guild_id) {
        let handler = handler_lock.lock().await;
        let queue = handler.queue();

        if queue.is_empty() {
            msg.reply(ctx, "Nothing is playing!").await?;
            return Ok(());
        }

       
        queue.skip()?;     //queue.skip() will skip the current track and start playing the next one in the queue

        let remaining = queue.len().saturating_sub(1);
        if remaining > 0 {
            msg.reply(ctx, format!("Skipped! {} track(s) remaining.", remaining)).await?;
        } else {
            msg.reply(ctx, "Skipped! Queue is now empty.").await?;
        }
    } else {
        msg.reply(ctx, "I'm not in a voice channel!").await?;
    }

    Ok(())
}