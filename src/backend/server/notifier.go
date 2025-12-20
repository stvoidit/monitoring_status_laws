package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"monitoring_draft_laws/internals/lawsparser"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (app *Application) sendBotMessage(ctx context.Context, chatID int64, text string) {
	if app.tgbot == nil {
		slog.Warn("ERROR app.tgbot is nil pointer")
		return
	}
	if _, err := app.tgbot.Send(tgbotapi.MessageConfig{
		BaseChat:              tgbotapi.BaseChat{ChatID: chatID, ReplyToMessageID: 0},
		Text:                  text,
		DisableWebPagePreview: true,
		ParseMode:             tgbotapi.ModeHTML,
	}); err != nil {
		errText := err.Error()
		slog.Error("sendBotMessage",
			slog.String("error", errText),
			slog.Int64("chatId", chatID))
		if strings.Contains(errText, "chat not found") {
			if err := app.db.DeleteChatID(ctx, chatID); err != nil {
				slog.Error("DeleteChatID",
					slog.String("error", err.Error()),
					slog.Int64("chatId", chatID))
			} else {
				slog.Info("DeleteChatID", slog.Int64("chatId", chatID))
			}
		}
	} else {
		slog.Info("sendBotMessage",
			slog.Int64("chatId", chatID),
			slog.String("text", text))
	}
}

func FilesToHTML(addr, source string, files []lawsparser.File) string {
	if len(files) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("<b>Новые файлы:</b>\n")
	for i, file := range files {
		var sourceURL = fmt.Sprintf("%s/api/proxy_download?%s", addr, file.QueryParams(source))
		sb.WriteString(fmt.Sprintf(`%d. <a href="%s">%s</a>`, i+1, sourceURL, file.Name))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}

// notify - нотификация пользователей в тг
func (app *Application) Notify(
	ctx context.Context,
	fd *lawsparser.FormatDocument,
	chatsID []int64,
	newStatus string,
	newFiles []lawsparser.File) (err error) {
	if len(chatsID) == 0 {
		return nil
	}
	if app.tgbot == nil {
		return errors.New("tgbot not inited")
	}
	if fd == nil {
		return errors.New("notification data is nil")
	}
	var addr = app.config.ServiceAddr()
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<b>Обновление в документе:</b> <a href="%s">%s</a>`, fd.OriginalHREF(), fd.DocumentID))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf(`<b>Название:</b> "%s"`, fd.Label))
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf(`<b>Краткое название:</b> "%s"`, fd.ShortLabel))
	sb.WriteString("\n\n")
	if len(newStatus) > 0 {
		sb.WriteString(fmt.Sprintf(`<b>Новый статус:</b> "%s"`, fd.CurrentStatus))
		sb.WriteString("\n\n")
	}
	if len(fd.Journal) > 0 {
		var lastEvent = fd.Journal[len(fd.Journal)-1]
		var lastEventDate, ok = lastEvent["date"]
		if ok {
			sb.WriteString(fmt.Sprintf(`<b>Дата последнего события:</b> %s`, lastEventDate))
			sb.WriteString("\n\n")
		}
	}
	sb.WriteString(FilesToHTML(addr, fd.SourceHost, newFiles))
	sb.WriteString(fmt.Sprintf(`<a href="%s/document?id=%s">перейти в карточку</a>`, addr, fd.DocumentID))
	tgmsg := sb.String()

	for i := range chatsID {
		app.sendBotMessage(ctx, chatsID[i], tgmsg)
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func (app *Application) readTgMessage(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		command := update.Message.Command()
		slog.Info("readTgMessage:",
			slog.Int64("chatID", update.Message.Chat.ID),
			slog.String("command", command),
			slog.String("text", update.Message.Text))
		switch command {
		case "start":
			decodeString, _ := base64.StdEncoding.DecodeString(strings.TrimSpace(strings.Replace(update.Message.Text, "/start", "", 1)))
			slog.Info("DecodeString", slog.String("command", command), slog.String("decodeString", string(decodeString)))
			userID, err := strconv.ParseUint(string(decodeString), 10, 64)
			if err != nil {
				slog.Error("ParseUint.UserID.TgBot",
					slog.Int64("chatID", update.Message.Chat.ID),
					slog.String("error", err.Error()))
				app.sendBotMessage(ctx,
					update.Message.Chat.ID,
					"Извините, ваш ID не распознан. Пожалуйста, воспользуйтесь кнопкой телеграм бота на сайте")
				break
			}
			username, err := app.db.CheckUserForBot(ctx, userID)
			if err != nil {
				slog.Error("CheckUserForBot",
					slog.Int64("chatID", update.Message.Chat.ID),
					slog.Uint64("userID", userID),
					slog.String("error", err.Error()))
				app.sendBotMessage(ctx, update.Message.Chat.ID,
					"Извините, ваш ID не распознан. Пожалуйста, воспользуйтесь кнопкой телеграм бота на сайте")
				break
			}
			if err := app.db.SaveChatID(ctx, update.Message.Chat.ID, userID, true); err == nil {
				app.sendBotMessage(ctx, update.Message.Chat.ID,
					fmt.Sprintf("Здравствуйте, %s\nУведомления по изменениям статуса документов включены.", username))
			} else {
				slog.Error("SaveChatID",
					slog.Int64("chatID", update.Message.Chat.ID),
					slog.String("error", err.Error()))
			}
		case "stop":
			if err := app.db.SaveChatID(ctx, update.Message.Chat.ID, 0, false); err != nil {
				slog.Error("SaveChatID:", slog.String("error", err.Error()))
			} else {
				slog.Info("SaveChatID", slog.String("command", command))
			}
		}
	case update.MyChatMember != nil:
		if update.MyChatMember.NewChatMember.Status == "kicked" {
			if err := app.db.SaveChatID(ctx, update.MyChatMember.Chat.ID, 0, false); err != nil {
				slog.Error("SaveChatID:",
					slog.Int64("chatID", update.Message.Chat.ID),
					slog.String("error", err.Error()))
			}
		}
	}
}

func (app *Application) initTgBot() (err error) {
	if app.config.TGBOT.TokenAPI == "" {
		return nil
	}
	app.tgbot, err = tgbotapi.NewBotAPI(app.config.TGBOT.TokenAPI)
	if err != nil {
		return err
	}
	me, err := app.tgbot.GetMe()
	if err != nil {
		return err
	}
	slog.Info("initTgBot", slog.Any("me", me))

	{
		response, err := app.tgbot.Request(tgbotapi.DeleteWebhookConfig{DropPendingUpdates: false})
		if err != nil {
			slog.Error("DeleteWebhookConfig", slog.String("error", err.Error()))
			return err
		} else {
			slog.Info("DeleteWebhookConfig", slog.Any("response", response))
		}
	}

	{
		commandsConfig := tgbotapi.NewSetMyCommands(
			tgbotapi.BotCommand{Command: "start", Description: "Начать получать уведомления"},
			tgbotapi.BotCommand{Command: "stop", Description: "Отключить получение уведомлений"},
		)
		commandsConfig.LanguageCode = "ru"
		response, err := app.tgbot.Request(commandsConfig)
		if err != nil {
			slog.Error("Send", slog.String("error", err.Error()))
			return err
		} else {
			slog.Info("NewSetMyCommands", slog.Any("response", response))
		}
	}
	return
}

func (app *Application) startTgPolling(ctx context.Context) (err error) {
	if app.tgbot == nil {
		return errors.New("tgbot not inited")
	}
	go func() {
		slog.Info("tgbot", slog.String("self", fmt.Sprintf("%+v", app.tgbot.Self)))
		pollConfig := tgbotapi.NewUpdate(0)
		pollConfig.Timeout = 60
		updates := app.tgbot.GetUpdatesChan(pollConfig)
		for {
			select {
			case <-ctx.Done():
				slog.Info("stop polling, context done")
				return
			case upd := <-updates:
				fmt.Println(upd)
				if upd.Message == nil {
					continue
				}
				app.readTgMessage(ctx, upd)
			}
		}
	}()
	return nil
}
