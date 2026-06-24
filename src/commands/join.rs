#[command]
async fn join(ctx: &Context, msg: &Message) -> CommandResult {
    let guild = msg.guild(&ctx.cache).unwrap();
    let channel_id = guild
        .voice_states
        .get(&msg.author.id)
        .and_then(|vs| vs.channel_id)
        .ok_or("You must be in a voice channel!")?;

    let manager = songbird::get(ctx).await.unwrap();
    let (_, success) = manager.join(guild.id, channel_id).await;
    success?;
    msg.reply(ctx, "Joined your voice channel!").await?;
    Ok(())
}