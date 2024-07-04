package main

import (
	"context"
	"errors"
	"fmt"
	telegramBot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kkdai/youtube/v2"
	"io"
	"log"
	"mime"
	"strings"
	"wezzaTubeBot/internal/youtubevideo"
)

const (
	startMessage   = "Welcome to WezzaTube Bot. Just send the youtube video url and I will download it and send it to you."
	invalidMessage = "The bot only understands messages that are a youtube video url. I don't understand anything else."
)

type bot struct {
	youtubeClient *youtube.Client
}

func newBot() *bot {
	return &bot{
		youtubeClient: &youtube.Client{},
	}
}

func (_ *bot) matchDownloadRequest(update *models.Update) bool {
	if update.Message == nil {
		return false
	}

	text := strings.TrimSpace(update.Message.Text)
	if text == "" {
		return false
	}

	if !youtubevideo.IsYouTubeVideoURL(text) {
		return false
	}

	return true
}

func (r *bot) downloadHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	var message *models.Message

	videoId, err := youtubevideo.ExtractYouTubeVideoID(strings.TrimSpace(update.Message.Text))
	if err != nil {
		message, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            err.Error(),
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}
		return
	}

	message, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
		ChatID:          update.Message.Chat.ID,
		Text:            "Trying to download the video...",
		ReplyParameters: replyParametersTo(update.Message),
	})
	if err != nil {
		log.Print(err)
	}

	var video *models.InputFileUpload
	video, err = r.download(videoId)
	if err != nil {
		log.Print(err)

		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          message.ID,
			Text:            err.Error(),
			ReplyParameters: replyParametersTo(message),
		})

		return
	}

	message, err = b.SendVideo(ctx, &telegramBot.SendVideoParams{
		ChatID:          update.Message.Chat.ID,
		Video:           video,
		Caption:         fmt.Sprintf("Here's your video..."),
		ReplyParameters: replyParametersTo(message),
	})
	if err != nil {
		log.Print(err)
	}
}

func (_ *bot) startHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if update.Message != nil {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            startMessage,
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}
	}
}

func (_ *bot) defaultHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	if update.Message != nil {
		_, err := b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            invalidMessage,
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}
	}
}

func (r *bot) download(videoID string) (*models.InputFileUpload, error) {
	video, err := r.youtubeClient.GetVideo(videoID)
	if err != nil {
		panic(fmt.Errorf("failed to download video: %w", err))
	}

	formats := video.Formats.WithAudioChannels()

	if len(formats) == 0 {
		return nil, errors.New("no formats found")
	}

	format := youtubevideo.HighestQualityFormat(formats)

	var stream io.ReadCloser
	stream, _, err = r.youtubeClient.GetStream(video, format)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}
	defer func(stream io.ReadCloser) {
		err = stream.Close()
		if err != nil {
			log.Print(fmt.Errorf("error closing stream: %w", err))
		}
	}(stream)

	var extensions []string
	extensions, err = mime.ExtensionsByType(format.MimeType)
	if err != nil {
		return nil, fmt.Errorf("error getting extensions: %w", err)
	}
	if len(extensions) == 0 {
		return nil, fmt.Errorf("no extension found for mime type %q", format.MimeType)
	}

	return &models.InputFileUpload{Filename: fmt.Sprintf("%s%s", video.Title, extensions[0]), Data: stream}, nil
}
