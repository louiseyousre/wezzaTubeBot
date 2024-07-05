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
	"regexp"
	"strings"
	"wezzaTubeBot/internal/youtubevideo"
)

const (
	startMessage   = "Welcome to WezzaTube Bot. Just send the youtube video url and I will downloadVideo it and send it to you."
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

	if !youtubevideo.IsYouTubeVideoURL(text) && !youtubevideo.IsYouTubePlaylistURL(text) {
		return false
	}

	return true
}

func (r *bot) downloadHandler(ctx context.Context, b *telegramBot.Bot, update *models.Update) {
	var message *models.Message
	trimmedText := strings.TrimSpace(update.Message.Text)

	if youtubevideo.IsYouTubePlaylistURL(trimmedText) {
		var err error
		message, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          update.Message.Chat.ID,
			Text:            "Trying to download the playlist...",
			ReplyParameters: replyParametersTo(update.Message),
		})
		if err != nil {
			log.Print(err)
		}

		ch := r.downloadPlaylist(ctx, trimmedText)
		for res := range ch {
			r.handleVideoDownloadingResult(ctx, b, update.Message.Chat.ID, message, res.FileUpload, res.Err)
		}
	} else {
		videoId, err := youtubevideo.ExtractYouTubeVideoID(trimmedText)
		if err != nil {
			_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
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
		video, err = r.downloadVideo(ctx, videoId)
		r.handleVideoDownloadingResult(ctx, b, update.Message.Chat.ID, message, video, err)
	}
}

func (r *bot) handleVideoDownloadingResult(ctx context.Context, b *telegramBot.Bot, chatId int64, repliedTo *models.Message, video *models.InputFileUpload, err error) {
	if err != nil {
		log.Print(err)

		_, err = b.SendMessage(ctx, &telegramBot.SendMessageParams{
			ChatID:          repliedTo.ID,
			Text:            err.Error(),
			ReplyParameters: replyParametersTo(repliedTo),
		})

		return
	}

	_, err = b.SendVideo(ctx, &telegramBot.SendVideoParams{
		ChatID:          chatId,
		Video:           video,
		Caption:         video.Filename,
		ReplyParameters: replyParametersTo(repliedTo),
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

var baseMimeRegex = regexp.MustCompile(`^\s*([a-zA-Z0-9\-\.]+/[a-zA-Z0-9\-\.]+)`)

func getExtensionForMimeType(mimeType string) (string, error) {
	match := baseMimeRegex.FindStringSubmatch(mimeType)

	if len(match) <= 1 {
		return "", errors.New("invalid mime type")
	}

	matchedMimeType := match[1]

	extensions, err := mime.ExtensionsByType(matchedMimeType)
	if err != nil {
		return "", fmt.Errorf("error getting extensions: %w", err)
	}
	if len(extensions) == 0 {
		return "", fmt.Errorf("no extension found for mime type %q", matchedMimeType)
	}

	return extensions[0], nil
}

type downloadResult struct {
	FileUpload *models.InputFileUpload
	Err        error
}

func (r *bot) downloadPlaylist(ctx context.Context, playlistUrl string) <-chan *downloadResult {
	resultCh := make(chan *downloadResult)

	go func() {
		defer close(resultCh)

		playlist, err := r.youtubeClient.GetPlaylistContext(ctx, playlistUrl)
		if err != nil {
			resultCh <- &downloadResult{Err: fmt.Errorf("error getting playlist: %w", err)}
			return
		}

		for _, video := range playlist.Videos {
			fileUpload, err := r.downloadVideo(ctx, video.ID)
			resultCh <- &downloadResult{FileUpload: fileUpload, Err: err}
		}
	}()

	return resultCh
}

func (r *bot) downloadVideo(ctx context.Context, videoID string) (*models.InputFileUpload, error) {
	video, err := r.youtubeClient.GetVideoContext(ctx, videoID)
	if err != nil {
		panic(fmt.Errorf("failed to downloadVideo video: %w", err))
	}

	formats := video.Formats.WithAudioChannels()

	if len(formats) == 0 {
		return nil, errors.New("no formats found")
	}

	format := youtubevideo.HighestQualityFormat(formats)

	var stream io.ReadCloser
	stream, _, err = r.youtubeClient.GetStreamContext(ctx, video, format)
	if err != nil {
		return nil, fmt.Errorf("failed to downloadVideo video: %w", err)
	}
	autoCloseStream := NewAutoCloseReadCloser(stream)

	var extension string
	extension, err = getExtensionForMimeType(format.MimeType)
	if err != nil {
		return nil, fmt.Errorf("error getting extension for mime type %q: %w", format.MimeType, err)
	}

	filename := fmt.Sprintf("%s%s", video.Title, extension)

	return &models.InputFileUpload{Filename: filename, Data: autoCloseStream}, nil
}

type AutoCloseReadCloser struct {
	io.ReadCloser
	done chan struct{}
}

func NewAutoCloseReadCloser(rc io.ReadCloser) *AutoCloseReadCloser {
	return &AutoCloseReadCloser{
		ReadCloser: rc,
		done:       make(chan struct{}),
	}
}

func (r *AutoCloseReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if err == io.EOF {
		close(r.done)
	}
	return n, err
}

func (r *AutoCloseReadCloser) Close() error {
	<-r.done // Wait until reading is done
	return r.ReadCloser.Close()
}
