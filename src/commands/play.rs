#[command]
async fn play(ctx: &Context, msg: &Message, args: Args) -> CommandResult {
    let url = args.rest().trim().to_string();
    if url.is_empty() {
        msg.reply(ctx, "Provide a URL: `!play <url>`").await?;
        return Ok(());
    }

    let guild_id = msg.guild_id.unwrap();
    let manager = songbird::get(ctx).await.unwrap();

    if let Some(handler_lock) = manager.get(guild_id) {
        let mut handler = handler_lock.lock().await;

        let source = songbird::input::YoutubeDl::new(
            reqwest::Client::new(),
            url,
        );

        handler.enqueue_input(source.into()).await;
        let queue_len = handler.queue().len();
if queue_len > 1 {
    msg.reply(ctx, format!("Added to queue! Position: {}", queue_len)).await?;
} else {
    msg.reply(ctx, "Now playing!").await?;
}
        msg.reply(ctx, "Now playing!").await?;
    } else {
        msg.reply(ctx, "Not in a voice channel. Use `!join` first.").await?;
    }
    Ok(())
}