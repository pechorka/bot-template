package events

import (
	"context"
	"encoding/json"
	"fmt"
	tbapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pechorka/bot-template/app/bot"
	"log"
)

type tbAPI interface {
	GetUpdatesChan(config tbapi.UpdateConfig) tbapi.UpdatesChannel
	Send(c tbapi.Chattable) (tbapi.Message, error)
	Request(c tbapi.Chattable) (*tbapi.APIResponse, error)
	GetChat(config tbapi.ChatInfoConfig) (tbapi.Chat, error)
}

type ibot interface {
	OnMessage(message bot.Message) bot.Response
}

type Listener struct {
	Bot ibot
	API tbAPI
}

// Do process all events, blocked call
func (l *Listener) Do(ctx context.Context) error {

	u := tbapi.NewUpdate(0)
	u.Timeout = 60

	updates := l.API.GetUpdatesChan(u)

	for {
		select {

		case <-ctx.Done():
			return ctx.Err()

		case update, ok := <-updates:
			if !ok {
				return fmt.Errorf("telegram update chan closed")
			}

			if update.Message == nil {
				log.Print("[DEBUG] empty message body")
				continue
			}

			msgJSON, errJSON := json.Marshal(update.Message)
			if errJSON != nil {
				log.Printf("[ERROR] failed to marshal update.Message to json: %v", errJSON)
				continue
			}
			log.Printf("[DEBUG] %s", string(msgJSON))

			if update.Message.Chat == nil {
				log.Print("[DEBUG] ignoring message not from chat")
				continue
			}

			fromChat := update.Message.Chat.ID

			msg := l.transform(update.Message)

			log.Printf("[DEBUG] incoming msg: %+v", msg)

			resp := l.Bot.OnMessage(*msg)

			if err := l.sendBotResponse(resp, fromChat); err != nil {
				log.Printf("[WARN] failed to respond on update, %v", err)
			}

		}
	}
}

// sendBotResponse sends bot's answer to tg channel and saves it to log
func (l *Listener) sendBotResponse(resp bot.Response, chatID int64) error {
	if !resp.Send {
		return nil
	}

	log.Printf("[DEBUG] bot response - %+v", resp.Text)
	tbMsg := tbapi.NewMessage(chatID, resp.Text)
	tbMsg.ParseMode = tbapi.ModeMarkdown
	_, err := l.API.Send(tbMsg)
	if err != nil {
		return fmt.Errorf("can't send message to telegram %q: %w", resp.Text, err)
	}

	return nil
}

func (l *Listener) transform(msg *tbapi.Message) *bot.Message {
	message := bot.Message{
		ID:   msg.MessageID,
		Sent: msg.Time(),
		Text: msg.Text,
	}

	if msg.Chat != nil {
		message.ChatID = msg.Chat.ID
	}

	if msg.From != nil {
		message.From = bot.User{
			ID:          msg.From.ID,
			Username:    msg.From.UserName,
			DisplayName: msg.From.FirstName + " " + msg.From.LastName,
		}
	}

	if msg.SenderChat != nil {
		message.SenderChat = bot.SenderChat{
			ID:       msg.SenderChat.ID,
			UserName: msg.SenderChat.UserName,
		}
	}

	switch {
	case msg.Entities != nil && len(msg.Entities) > 0:
		message.Entities = l.transformEntities(msg.Entities)

	case msg.Photo != nil && len(msg.Photo) > 0:
		sizes := msg.Photo
		lastSize := sizes[len(sizes)-1]
		message.Image = &bot.Image{
			FileID:   lastSize.FileID,
			Width:    lastSize.Width,
			Height:   lastSize.Height,
			Caption:  msg.Caption,
			Entities: l.transformEntities(msg.CaptionEntities),
		}
	}

	// fill in the message's reply-to message
	if msg.ReplyToMessage != nil {
		message.ReplyTo.Text = msg.ReplyToMessage.Text
		message.ReplyTo.Sent = msg.ReplyToMessage.Time()
		if msg.ReplyToMessage.From != nil {
			message.ReplyTo.From = bot.User{
				ID:          msg.ReplyToMessage.From.ID,
				Username:    msg.ReplyToMessage.From.UserName,
				DisplayName: msg.ReplyToMessage.From.FirstName + " " + msg.ReplyToMessage.From.LastName,
			}
		}
		if msg.ReplyToMessage.SenderChat != nil {
			message.ReplyTo.SenderChat = bot.SenderChat{
				ID:       msg.ReplyToMessage.SenderChat.ID,
				UserName: msg.ReplyToMessage.SenderChat.UserName,
			}
		}
	}

	return &message
}

func (l *Listener) transformEntities(entities []tbapi.MessageEntity) *[]bot.Entity {
	if len(entities) == 0 {
		return nil
	}

	var result []bot.Entity
	for _, entity := range entities {
		e := bot.Entity{
			Type:   entity.Type,
			Offset: entity.Offset,
			Length: entity.Length,
			URL:    entity.URL,
		}
		if entity.User != nil {
			e.User = &bot.User{
				ID:          entity.User.ID,
				Username:    entity.User.UserName,
				DisplayName: entity.User.FirstName + " " + entity.User.LastName,
			}
		}
		result = append(result, e)
	}

	return &result
}
