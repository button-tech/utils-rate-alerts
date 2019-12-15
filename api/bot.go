package api

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"strings"
)

const (
	help      = `/alert <price> <currency> <fiat> <condition> - i.e. /alert 7500 btc usd <`
	errorMsg  = `sorry, we have an error. try it latter.`
	okMsg     = `you subscribed. Wait your alert :)`
	resultMsg = `Your condition: %s
price: %s
current price : %s
currency: %s
fiat: %s 
done ✅.`
)

type BotProvider struct {
	Channel  *amqp.Channel
	Queue    amqp.Queue
	BotToken string
}

type Bot struct {
	api       *tgbotapi.BotAPI
	tgChannel tgbotapi.UpdatesChannel

	channel *amqp.Channel
	queue   amqp.Queue
}

type botMessage struct {
	chatID    int64
	messageID int
	userText  string
}

type arguments struct {
	price     string
	condition string
	fiat      string
	currency  string
	chatID    int64
}

func CreateBot(p BotProvider) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(p.BotToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 10
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}
	return &Bot{
		api:       bot,
		tgChannel: updates,
		channel:   p.Channel,
		queue:     p.Queue,
	}, nil
}

func SetupBot(c *amqp.Channel, q amqp.Queue, t string) BotProvider {
	return BotProvider{
		Channel:  c,
		Queue:    q,
		BotToken: t,
	}
}

func (b *Bot) Processing() {
	for update := range b.tgChannel {
		if update.Message == nil {
			continue
		}

		var msgConfig tgbotapi.MessageConfig
		bmsg := botMsg(update.Message.Chat.ID, update.Message.MessageID, update.Message.Text)
		args := update.Message.CommandArguments()
		if !ifArgs(args) {
			msgConfig = bmsg.helper()
			if _, err := b.api.Send(msgConfig); err != nil {
				log.Println(err)
			}
			continue
		}

		switch update.Message.Command() {
		case "alert":
			argsToSubscribe := splitArgs(args, bmsg.chatID)
			if err := b.subscribeUser(argsToSubscribe); err != nil {
				msgConfig = bmsg.messageToUser(errorMsg)
			} else {
				msgConfig = bmsg.messageToUser(okMsg)
			}
		default:
			msgConfig = bmsg.helper()
		}

		if _, err := b.api.Send(msgConfig); err != nil {
			log.Println(err)
		}
	}
}

func ifArgs(s string) bool {
	return s != ""
}

func botMsg(chatID int64, msgID int, text string) *botMessage {
	return &botMessage{
		chatID:    chatID,
		messageID: msgID,
		userText:  text,
	}
}

func (b *Bot) AlertUser(c trueCondition) error {
	alertMsg := fmt.Sprintf(
		resultMsg,
		c.Values.Condition,
		c.Values.Price,
		c.Values.CurrentPrice,
		c.Values.Currency,
		c.Values.Fiat,
	)
	chatID, err := strconv.ParseInt(c.URL, 10, 64)
	if err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(chatID, alertMsg)
	_, err = b.api.Send(msg)
	return err
}

func (b *Bot) subscribeUser(args alert) error {
	body, err := json.Marshal(&args)
	if err != nil {
		return err
	}

	err = b.channel.Publish(
		"",
		b.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	return err
}

func (bmsg *botMessage) messageToUser(text string) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(bmsg.chatID, text)
}

func (bmsg *botMessage) helper() tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(bmsg.chatID, help)
}

func splitArgs(args string, chatID int64) alert {
	splitted := strings.Split(args, " ")
	convChatID := strconv.FormatInt(chatID, 10)
	return alert{
		Price:     splitted[0],
		Currency:  splitted[1],
		Fiat:      splitted[2],
		Condition: splitted[3],
		URL:       convChatID,
	}
}
