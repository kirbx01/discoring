use serenity::framework::standard::{macros::command, Args, CommandResult};
use serenity::model::prelude::*;
use serenity::prelude::*;

#[command]
#[only_in(guilds)]
async fn leave(ctx: &Context, msg: &Message, _args: Args) -> CommandResult {
    let guild_id = msg.guild_id.unwrap();
    let manager = songbird::get(ctx).await.unwrap();

    if manager.get(guild_id).is_some() {
        
        if let Some(handler_lock) = manager.get(guild_id) { // stop queue first, then disconnect
            handler_lock.lock().await.queue().stop();
        }

        manager.remove(guild_id).await?;
        msg.reply(ctx, "Left and cleared the queue!").await?;
    } else {
        msg.reply(ctx, "I'm not in a voice channel!").await?;
    }

    Ok(())
}