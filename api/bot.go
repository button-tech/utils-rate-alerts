package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/streadway/amqp"
)

const (
	helpMessage        = `/alert - for use`
	errorMsg           = `❌ Произошла ошибка. Попробуйте позже`
	alertMessage       = `✅ Вы подписаны на уведомление`
	alertResultMessage = `ℹ️ Уведомление о %s
Текущая цена: %s %s
Условие выполнения: %s %s %s`
)

const (
	firstPage = `Выберите крипто валюту:
Пример: BTC 
`
	secondPage = `Выберите фиатную валюту:
Пример: USD
`
	thirdPage = `Введите сумму в %s:
Пример: 7000`
	fourthPage = `Введите условие:
Пример: <= или >= или == или < или >
`
)

const (
	errCryptoInput    = "❌ Попробуйте другую крипто валюту\nПример: BTC"
	errFiatInput      = "❌ Попробуйте другую фиатную валюту\nПример: USD"
	errPriceInput     = "❌ Введите валидную сумму\nПример: 7000"
	errConditionInput = "❌ Введите доступное условие\nПример: <= или >= или == или < или >"
)

var fiats = map[string]struct{}{"AED": struct{}{}, "ALL": struct{}{}, "AMD": struct{}{}, "AOA": struct{}{}, "ARS": struct{}{}, "AUD": struct{}{}, "BAM": struct{}{}, "BDT": struct{}{}, "BGN": struct{}{}, "BHD": struct{}{}, "BIF": struct{}{}, "BND": struct{}{}, "BOB": struct{}{}, "BRL": struct{}{}, "BSD": struct{}{}, "BTC": struct{}{}, "BTN": struct{}{}, "BWP": struct{}{}, "BYN": struct{}{}, "CAD": struct{}{}, "CDF": struct{}{}, "CHF": struct{}{}, "CLP": struct{}{}, "CNY": struct{}{}, "COP": struct{}{}, "CRC": struct{}{}, "CZK": struct{}{}, "DKK": struct{}{}, "DOP": struct{}{}, "DZD": struct{}{}, "EGP": struct{}{}, "ETB": struct{}{}, "EUR": struct{}{}, "GBP": struct{}{}, "GEL": struct{}{}, "GGP": struct{}{}, "GHS": struct{}{}, "GIP": struct{}{}, "GTQ": struct{}{}, "HKD": struct{}{}, "HNL": struct{}{}, "HRK": struct{}{}, "HUF": struct{}{}, "IDR": struct{}{}, "ILS": struct{}{}, "INR": struct{}{}, "IQD": struct{}{}, "IRR": struct{}{}, "ISK": struct{}{}, "JMD": struct{}{}, "JOD": struct{}{}, "JPY": struct{}{}, "KES": struct{}{}, "KGS": struct{}{}, "KHR": struct{}{}, "KRW": struct{}{}, "KWD": struct{}{}, "KZT": struct{}{}, "LBP": struct{}{}, "LKR": struct{}{}, "LSL": struct{}{}, "MAD": struct{}{}, "MDL": struct{}{}, "MMK": struct{}{}, "MOP": struct{}{}, "MUR": struct{}{}, "MWK": struct{}{}, "MXN": struct{}{}, "MYR": struct{}{}, "NAD": struct{}{}, "NGN": struct{}{}, "NIO": struct{}{}, "NOK": struct{}{}, "NPR": struct{}{}, "NZD": struct{}{}, "OMR": struct{}{}, "PAB": struct{}{}, "PEN": struct{}{}, "PGK": struct{}{}, "PHP": struct{}{}, "PKR": struct{}{}, "PLN": struct{}{}, "PYG": struct{}{}, "QAR": struct{}{}, "RON": struct{}{}, "RUB": struct{}{}, "RWF": struct{}{}, "SAR": struct{}{}, "SBD": struct{}{}, "SEK": struct{}{}, "SGD": struct{}{}, "SHP": struct{}{}, "SZL": struct{}{}, "THB": struct{}{}, "TMT": struct{}{}, "TND": struct{}{}, "TOP": struct{}{}, "TRY": struct{}{}, "TTD": struct{}{}, "TWD": struct{}{}, "TZS": struct{}{}, "UAH": struct{}{}, "UGX": struct{}{}, "USD": struct{}{}, "UYU": struct{}{}, "UZS": struct{}{}, "VEF": struct{}{}, "VND": struct{}{}, "VUV": struct{}{}, "XAF": struct{}{}, "XAU": struct{}{}, "XCD": struct{}{}, "XOF": struct{}{}, "ZAR": struct{}{}, "ZMW": struct{}{}}
var cryptoCurrencies = map[string]struct{}{"ADA": struct{}{}, "AE": struct{}{}, "ALGO": struct{}{}, "ARDR": struct{}{}, "ATOM": struct{}{}, "BCD": struct{}{}, "BCH": struct{}{}, "BCN": struct{}{}, "BNB": struct{}{}, "BSV": struct{}{}, "BTC": struct{}{}, "BTG": struct{}{}, "BTM": struct{}{}, "BTS": struct{}{}, "BTT": struct{}{}, "CENNZ": struct{}{}, "DASH": struct{}{}, "DCR": struct{}{}, "DGB": struct{}{}, "DOGE": struct{}{}, "EOS": struct{}{}, "ETC": struct{}{}, "ETH": struct{}{}, "ICX": struct{}{}, "IOST": struct{}{}, "KMD": struct{}{}, "LSK": struct{}{}, "LTC": struct{}{}, "LUNA": struct{}{}, "MONA": struct{}{}, "NANO": struct{}{}, "NEO": struct{}{}, "NRG": struct{}{}, "ONT": struct{}{}, "QTUM": struct{}{}, "RVN": struct{}{}, "STEEM": struct{}{}, "STRAT": struct{}{}, "THETA": struct{}{}, "TOMO": struct{}{}, "TRX": struct{}{}, "VET": struct{}{}, "VSYS": struct{}{}, "WAVES": struct{}{}, "XEM": struct{}{}, "XLM": struct{}{}, "XMR": struct{}{}, "XRP": struct{}{}, "XTZ": struct{}{}, "XVG": struct{}{}, "ZEC": struct{}{}, "ZEN": struct{}{}, "ZIL": struct{}{}}

var conditionsVarifier = map[string]struct{}{
	">":  {},
	"<":  {},
	"==": {},
	">=": {},
	"<=": {},
}

type BotProvider struct {
	Channel  *amqp.Channel
	Queue    amqp.Queue
	BotToken string
}

type Bot struct {
	api       *tgbotapi.BotAPI
	tgChannel tgbotapi.UpdatesChannel
	cache     *cache

	channel *amqp.Channel
	queue   amqp.Queue
}

type cache struct {
	mu          sync.Mutex
	subscribers map[int64][]page
}

type page struct {
	number    int
	userInput string
}

func newCache() *cache {
	return &cache{
		mu:          sync.Mutex{},
		subscribers: make(map[int64][]page),
	}
}

func (c *cache) set(k int64, p page) []page {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.subscribers[k]
	if !ok {
		c.subscribers[k] = make([]page, 0, 4)
	}
	val = append(val, p)
	c.subscribers[k] = val
	return val
}

func (c *cache) get(k int64) (ps []page, ok bool) {
	if ps, ok = c.subscribers[k]; !ok {
		return nil, false
	}
	return
}

func (c *cache) back(k int64) {
	c.mu.Lock()
	ps := c.subscribers[k]
	if len(ps) < 1 {
		return
	}
	c.subscribers[k] = ps[:len(ps)-1]
	c.mu.Unlock()
}

func (c *cache) delete(k int64) {
	c.mu.Lock()
	delete(c.subscribers, k)
	c.mu.Unlock()
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
		cache:     newCache(),
	}, nil
}

func SetupBot(c *amqp.Channel, q amqp.Queue, t string) BotProvider {
	return BotProvider{
		Channel:  c,
		Queue:    q,
		BotToken: t,
	}
}

func (p *page) giveContent(page int) (c string) {
	switch page {
	case 0:
		c = firstPage
	case 1:
		c = secondPage
	case 2:
		c = fmt.Sprintf(thirdPage, strings.ToUpper(p.userInput))
	case 3:
		c = fourthPage
	}
	return
}

func (b *Bot) Processing(ctx context.Context, wg *sync.WaitGroup) {
	for update := range b.tgChannel {
		select {
		case <-ctx.Done():
			log.Println("receive bots updates shutdown")
			wg.Done()
			return
		default:
			if update.Message == nil {
				continue
			}

			if update.CallbackQuery != nil {
				fmt.Print(update)

				b.api.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data))

				b.api.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data))
			}

			chatID := update.Message.Chat.ID
			userText := update.Message.Text
			switch update.Message.Command() {
			case "alert":
				if _, ok := b.cache.get(chatID); !ok {
					b.cache.delete(chatID)
				}
				p := page{userInput: "", number: 0}
				b.cache.set(chatID, p)
				text := p.giveContent(0)
				msg := tgbotapi.NewMessage(chatID, text)
				msg.ReplyMarkup = backKeyBoard
				b.api.Send(msg)
				continue
			case "help":
				b.api.Send(help(chatID))
				continue
			}

			pages, ok := b.cache.get(chatID)
			if !ok {
				b.api.Send(help(chatID))
				continue
			}

			if userText == "back" {
				var text string
				if len(pages) == 4 {
					text = pages[len(pages)-2].giveContent(len(pages) - 2)
				} else {
					p := page{}
					text = p.giveContent(len(pages) - 2)
				}
				msg := tgbotapi.NewMessage(chatID, text)
				b.cache.back(chatID)
				b.api.Send(msg)
				continue
			}

			if pages[len(pages)-1].number == 0 {
				_, ok := cryptoCurrencies[strings.ToUpper(userText)]
				if !ok {
					//p := page{}
					//text := p.giveContent(len(pages) - 1)
					msg := tgbotapi.NewMessage(chatID, errCryptoInput)
					b.api.Send(msg)
					continue
				}
			}

			if pages[len(pages)-1].number == 1 {
				_, ok := fiats[strings.ToUpper(userText)]
				if !ok {
					//p := page{}
					//text := p.giveContent(len(pages) - 1)
					msg := tgbotapi.NewMessage(chatID, errFiatInput)
					b.api.Send(msg)
					continue
				}
			}

			if pages[len(pages)-1].number == 2 {
				_, err := strconv.ParseFloat(userText, 10)
				if err != nil {
					b.api.Send(tgbotapi.NewMessage(chatID, errPriceInput))
					continue
				}
			}

			if pages[len(pages)-1].number == 3 {
				if _, ok := conditionsVarifier[userText]; !ok {
					b.api.Send(tgbotapi.NewMessage(chatID, errConditionInput))
					continue
				}
			}

			p := page{
				userInput: userText,
				number:    pages[len(pages)-1].number + 1,
			}
			pages = b.cache.set(chatID, p)
			text := p.giveContent(len(pages) - 1)
			msg := tgbotapi.NewMessage(chatID, text)
			if len(pages) == 4 {
				msg.ReplyMarkup = numericKeyboard
			}
			b.api.Send(msg)

			if len(pages) == 5 {
				alert := splitArgs(pages[1:], chatID)
				var text string
				if err := b.subscribeUser(alert); err != nil {
					text = errorMsg
				} else {
					text = alertMessage
				}
				b.api.Send(tgbotapi.NewMessage(chatID, text))
				b.cache.delete(chatID)
			}
		}
	}
}

func help(id int64) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(id, helpMessage)
}

func (b *Bot) AlertUser(c trueCondition) error {
	uc := strings.ToUpper(c.Values.Currency)
	alertMsg := fmt.Sprintf(
		alertResultMessage,
		uc,
		c.Values.CurrentPrice,
		strings.ToUpper(c.Values.Fiat),
		uc,
		c.Values.Condition,
		c.Values.Price,
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

func splitArgs(args []page, chatID int64) alert {
	convChatID := strconv.FormatInt(chatID, 10)
	return alert{
		Currency:  args[0].userInput,
		Fiat:      args[1].userInput,
		Price:     args[2].userInput,
		Condition: args[3].userInput,
		URL:       convChatID,
	}
}

var backKeyBoard = tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
	tgbotapi.NewKeyboardButton("back")),
)

var conditionsKeyBoard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(">="),
		tgbotapi.NewKeyboardButton("=="),
		tgbotapi.NewKeyboardButton("<=")),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(">"),
		tgbotapi.NewKeyboardButton("<"),
	),
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
		tgbotapi.NewInlineKeyboardButtonSwitch("2sw", "open 2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
)

// Cryptos:
//[HKD HNL BWP XOF PHP BHD KES BTC IQD IRR NIO ARS ETB RUB SZL AUD XCD KHR CHF NPR RON SEK KZT CRC XAU TWD USD BSD AED GBP OMR BRL CLP GIP JPY BYN NOK NGN TOP CDF PGK DZD PYG TND KWD LBP COP BIF PAB TTD TMT ISK BND VEF SAR KRW PLN SBD PEN HUF AMD MOP NZD GHS BAM XAF EUR EGP LSL MDL THB TRY ZMW QAR INR BOB GEL BTN CNY GTQ UAH UYU VUV MWK MAD KGS LKR MYR ILS ZAR VND HRK UGX UZS MUR CZK MXN PKR JOD AOA MMK DKK IDR BGN SHP JMD NAD TZS ALL CAD DOP GGP BDT RWF SGD]

// Fiats:
//[XAF XAU XCD XOF ZAR UZS VEF VUV ZMW VND]
//[CAD CNY CRC GBP CDF CLP COP EGP ETB DKK CHF CZK DOP DZD EUR]
//[SEK SZL THB TND TOP SGD TRY TTD UAH UGX USD UYU SHP TMT TWD TZS]
//[MOP KGS KRW KWD LBP LSL MAD MMK MUR KHR KZT LKR MDL]
//[BTC BDT BAM BGN BOB BRL AED AMD ARS AUD BHD BIF BND BSD ALL BYN BTN BWP AOA]
//[NAD SBD PLN MXN NOK NZD PAB PGK PHP PKR QAR MYR NGN NIO NPR OMR PEN RUB MWK PYG RON RWF SAR]
//[GGP GHS HKD JMD GEL GTQ HUF INR KES HNL ILS IQD IRR JOD GIP HRK IDR ISK JPY]
