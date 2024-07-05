package main

import (
	"context"
	telegramBot "github.com/go-telegram/bot"
	"log"
	"mime"
	"os"
	"os/signal"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go startHealthCheckServer(ctx)

	b := newBot()

	opts := []telegramBot.Option{
		telegramBot.WithDefaultHandler(b.defaultHandler),
	}

	tgBot, err := telegramBot.New(os.Getenv("TELEGRAM_BOT_TOKEN"), opts...)
	if err != nil {
		log.Fatal(err)
	}

	tgBot.RegisterHandler(
		telegramBot.HandlerTypeMessageText,
		"/start",
		telegramBot.MatchTypeExact,
		b.startHandler,
	)

	tgBot.RegisterHandlerMatchFunc(b.matchDownloadRequest, b.downloadHandler)

	log.Print("Starting the bot server...")

	tgBot.Start(ctx)

}

func init() {
	mimeTypes := map[string]string{
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".flv":  "video/x-flv",
		".3gp":  "video/3gpp",
		".mov":  "video/quicktime",
	}

	// Add each MIME type
	for ext, typ := range mimeTypes {
		err := mime.AddExtensionType(ext, typ)
		if err != nil {
			log.Fatalf("Failed to add MIME type for %s: %v", ext, err)
		}
		log.Printf("Added MIME type %s for extension %s", typ, ext)
	}
}
