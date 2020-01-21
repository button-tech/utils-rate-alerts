package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"log"
	"strconv"
	"strings"
	"sync"

	processCache "github.com/button-tech/utils-rate-alerts/pkg/storage/cache"
	t "github.com/button-tech/utils-rate-alerts/types"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
	"os"
)

type userAlert struct {
	currency  string
	condition string
	price     string
	fiat      string
}

type page struct {
	number    int
	userInput string
}

func keyGenForAlert(chatID int64) string {
	convChatID := strconv.FormatInt(chatID, 10)
	return fmt.Sprintf("%s_%s", convChatID, "alerts")
}

func keyGen(k int64) string {
	return strconv.FormatInt(k, 10)
}

func SetupBot(c *amqp.Channel, q amqp.Queue, t string) BotProvider {
	return BotProvider{
		Channel:  c,
		Queue:    q,
		BotToken: t,
	}
}

type BotProvider struct {
	Channel      *amqp.Channel
	Queue        amqp.Queue
	BotToken     string
	ProcessCache *processCache.Cache
}

type Bot struct {
	api                 *tgbotapi.BotAPI
	tgChannel           tgbotapi.UpdatesChannel
	cache               *cache
	deleteProcessingURL string
	channel             *amqp.Channel
	queue               amqp.Queue
}

func (b *Bot) AlertUser(c t.TrueCondition) error {
	userSettings := strings.Split(c.URL, "_")
	alertMsg := completedAlertMessage(c, userSettings[1])
	chatID, err := strconv.ParseInt(userSettings[0], 10, 64)
	if err != nil {
		return err
	}
	value := genAlertValue(c.Values.Currency, c.Values.Fiat, c.Values.Price, c.Values.Condition)
	alerts, _ := b.cache.getRawAlerts(chatID)
	for i, a := range alerts {
		if a == value {
			alerts = append(alerts[:i], alerts[i+1:]...)
			b.cache.setRawAlerts(chatID, alerts)
		}
	}

	msg := tgbotapi.NewMessage(chatID, alertMsg)
	_, err = b.api.Send(msg)
	return err
}

func (b *Bot) subscribeUser(args t.Alert) error {
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

// currency, fiat, price, condition
// [Token(b.Currency)][Fiat(b.Fiat)][URL(b.URL
func (b *Bot) deleteFromProcessCache(chatID int64, language, alert string) error {
	convChatID := strconv.FormatInt(chatID, 10)
	url := fmt.Sprintf("%s_%s", convChatID, language)
	splitted := strings.Split(alert, "_")
	block := processCache.ConditionBlock{
		Currency: splitted[0],
		Fiat:     splitted[1],
		URL:      url,
	}

	resp, err := req.Post(b.deleteProcessingURL+"delete", req.BodyJSON(&block))
	if err != nil {
		return err
	}

	if resp.Response().StatusCode != 201 {
		return errors.New("response status: not ok")
	}

	return nil

}

func (b *Bot) ProcessingUpdates(ctx context.Context, wg *sync.WaitGroup) {
	for update := range b.tgChannel {
		select {
		case <-ctx.Done():
			log.Println("receive bots updates shutdown")
			wg.Done()
			return
		default:
			var language string
			if update.CallbackQuery != nil {
				chatID := update.CallbackQuery.Message.Chat.ID
				if _, ok := b.cache.get(chatID); ok {
					b.cache.delete(chatID)
				}
				username := update.CallbackQuery.Message.Chat.UserName
				language = b.cache.setupLanguage(username, update.CallbackQuery.Data)
				msg := help(chatID, language)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			}
			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			var ok bool
			ok, language = b.cache.checkLanguage(update.Message.Chat.UserName)
			if !ok {
				language = "english"
			}
			userText := update.Message.Text
			switch update.Message.Command() {
			case "alert":
				if _, ok := b.cache.get(chatID); ok {
					b.cache.delete(chatID)
				}
				p := page{userInput: "", number: 0}
				b.cache.set(chatID, p)
				text := p.giveContent(0, language)
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = backKeyboard(language)
				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			case "help":
				if _, err := b.api.Send(help(chatID, language)); err != nil {
					log.Println(err)
				}
				continue
			case "language":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, selectLanguageMsg(language))
				msg.ReplyMarkup = languagesKeyBoard
				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			case "alerts":
				alerts, ok := b.cache.getAlerts(chatID)
				var text string
				if !ok || alerts == "" {
					text = selectNoAlertMessage(language)
				} else {
					text = alerts
				}
				msg := tgbotapi.NewMessage(chatID, text)
				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			case "delete":
				cmd := update.Message.CommandArguments()
				msg := tgbotapi.NewMessage(chatID, "")
				check, err := strconv.ParseFloat(cmd, 10)
				if cmd == "" || err != nil || check <= 0 {
					msg.Text = selectAlertNumber(language)
					if _, err := b.api.Send(msg); err != nil {
						log.Println(err)
					}
					continue
				}

				deleted, ok := b.cache.deleteAlert(chatID, cmd)
				if !ok {
					msg = tgbotapi.NewMessage(chatID, selectAlertNumber(language))
				} else {
					alerts, ok := b.cache.getAlerts(chatID)
					var text string
					if !ok {
						text = selectNoAlertMessage(language)
					} else {
						if alerts == "" {
							text = selectNoAlertMessage(language)
						} else {
							text = alerts
						}
					}
					if deleted != "" {
						if err := b.deleteFromProcessCache(chatID, language, deleted); err != nil {
							text = selectNoAlertMessage(language)
						}
					}
					msg.Text = text
				}

				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			}

			pages, ok := b.cache.get(chatID)
			if !ok {
				if _, err := b.api.Send(help(chatID, language)); err != nil {
					log.Println(err)
				}
				continue
			}

			if (userText == "back" && language == "english") || (userText == "назад" && language == "russian") {
				var text string
				if len(pages) == 4 {
					text = pages[len(pages)-2].giveContent(len(pages)-2, language)
				} else {
					p := page{}
					text = p.giveContent(len(pages)-2, language)

				}
				msg := tgbotapi.NewMessage(chatID, text)
				if len(pages) == 4 {
					msg.ReplyMarkup = backKeyboard(language)
				}
				if len(pages) <= 1 {
					msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}
				b.cache.back(chatID)
				if _, err := b.api.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			}

			if pages[len(pages)-1].number == 0 {
				_, ok := cryptoCurrencies[strings.ToUpper(userText)]
				if !ok {
					msg := tgbotapi.NewMessage(chatID, handleErrorInput(0, language))
					if _, err := b.api.Send(msg); err != nil {
						log.Println(err)
					}
					continue
				}
			}

			if pages[len(pages)-1].number == 1 {
				_, ok := fiats[strings.ToUpper(userText)]
				if !ok {
					msg := tgbotapi.NewMessage(chatID, handleErrorInput(1, language))
					if _, err := b.api.Send(msg); err != nil {
						log.Println(err)
					}
					continue
				}
			}

			if pages[len(pages)-1].number == 2 {
				_, err := strconv.ParseFloat(userText, 10)
				if err != nil {
					if _, err := b.api.Send(tgbotapi.NewMessage(chatID, handleErrorInput(2, language))); err != nil {
						log.Println(err)
					}
					continue
				}
			}

			if pages[len(pages)-1].number == 3 {
				if _, ok := conditionsVerifier[userText]; !ok {
					if _, err := b.api.Send(tgbotapi.NewMessage(chatID, handleErrorInput(3, language))); err != nil {
						log.Println(err)
					}
					continue
				}
			}

			p := page{
				userInput: userText,
				number:    pages[len(pages)-1].number + 1,
			}
			pages = b.cache.set(chatID, p)
			var msg tgbotapi.MessageConfig

			if pages[len(pages)-1].number == 4 {
				alert := splitArgs(pages[1:], chatID, language)
				var text string
				b.cache.setAlert(
					chatID,
					pages[len(pages)-4].userInput,
					pages[len(pages)-3].userInput,
					pages[len(pages)-2].userInput,
					pages[len(pages)-1].userInput,
				)

				if err := b.subscribeUser(alert); err != nil {
					text = handleErrorInput(4, language)
				} else {
					text = alertMessage(language)
				}
				msg = tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				b.cache.delete(chatID)
			} else {
				text := p.giveContent(len(pages)-1, language)
				msg = tgbotapi.NewMessage(chatID, text)
				if len(pages) == 4 {
					msg.ReplyMarkup = conditionsKeyboard(language)
				}
			}
			if _, err := b.api.Send(msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func CreateBot(p BotProvider) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(p.BotToken)
	if err != nil {
		return nil, err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:                 bot,
		tgChannel:           updates,
		channel:             p.Channel,
		queue:               p.Queue,
		cache:               newCache(),
		deleteProcessingURL: os.Getenv("PROCESSING_API_URL"),
	}, nil
}

func completedAlertMessage(c t.TrueCondition, language string) string {
	var format string
	uc := strings.ToUpper(c.Values.Currency)
	switch language {
	case "english":
		format = alertResultMessageENG
	case "russian":
		format = alertResultMessageRUS
	}
	return fmt.Sprintf(
		format,
		uc,
		c.Values.CurrentPrice,
		strings.ToUpper(c.Values.Fiat),
		uc,
		c.Values.Condition,
		c.Values.Price,
	)
}

func splitArgs(args []page, chatID int64, language string) t.Alert {
	convChatID := strconv.FormatInt(chatID, 10)
	l := fmt.Sprintf("%s_%s", convChatID, language)
	return t.Alert{
		Currency:  args[0].userInput,
		Fiat:      args[1].userInput,
		Price:     args[2].userInput,
		Condition: args[3].userInput,
		URL:       l,
	}
}
