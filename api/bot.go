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
	helpMessageENG     = `/alert - to sign up for a currency notice`
	helpMessageRUS     = `/alert - подписаться на уведомление по валюте`
	languageMessageRUS = "Выберите язык"
	languageMessageENG = "Select language"
	alertResultMessage = `ℹ️ Уведомление о %s
Текущая цена: %s %s
Условие выполнения: %s %s %s`
)

const (
	firstPageRUS = `Выберите крипто валюту:
Пример: BTC 
`
	secondPageRUS = `Выберите фиатную валюту:
Пример: USD
`
	thirdPageRUS = `Введите сумму в %s:
Пример: 7000`
	fourthPageRUS = `Введите условие:
Пример: <= или >= или == или < или >
`
	firstPageENG = `Select crypto currency:
Example: BTC 
`
	secondPageENG = `Select fiat currency:
Example: USD
`
	thirdPageENG = `Enter the amount in %s:
Example: 7000`
	fourthPageENG = `Enter the condition:
Example: <= или >= или == или < или >
`
)

const (
	errCryptoInputRUS    = "❌ Попробуйте другую крипто валюту\nПример: BTC"
	errFiatInputRUS      = "❌ Попробуйте другую фиатную валюту\nПример: USD"
	errPriceInputRUS     = "❌ Введите валидную сумму\nПример: 7000"
	errConditionInputRUS = "❌ Введите доступное условие\nПример: <= или >= или == или < или >"
	errAlertMsgRUS       = `❌ Произошла ошибка. Попробуйте позже`
	alertMessageRUS      = `✅ Вы подписаны на уведомление`

	errCryptoInputENG    = "❌ Try another crypto currency\nExample: BTC"
	errFiatInputENG      = "❌ Try another fiat currency\nExample: USD"
	errPriceInputENG     = "❌ Enter valid amount\nExample: 7000"
	errConditionInputENG = "❌ Enter an available condition\nExample: <= или >= или == или < или >"
	errAlertMsgENG       = `❌ An error has occurred. try late`
	alertMessageENG      = `✅ You subscribed to the notification`
)

func selectLanguageMsg(language string) (l string) {
	switch language {
	case "russian":
		l = languageMessageRUS
	case "english":
		l = languageMessageENG
	}
	return
}

func handleErrorInput(page int64, language string) (err string) {
	switch language {
	case "russian":
		switch page {
		case 0:
			err = errCryptoInputRUS
		case 1:
			err = errFiatInputRUS
		case 2:
			err = errPriceInputRUS
		case 3:
			err = errConditionInputRUS
		case 4:
			err = errAlertMsgRUS
		}
	case "english":
		switch page {
		case 0:
			err = errCryptoInputENG
		case 1:
			err = errFiatInputENG
		case 2:
			err = errPriceInputENG
		case 3:
			err = errConditionInputENG
		case 4:
			err = errAlertMsgENG
		}
	}
	return err
}

func alertMessage(language string) (n string) {
	switch language {
	case "russian":
		n = alertMessageRUS
	case "english":
		n = alertMessageENG
	}
	return
}

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
	subscribers map[string][]page
	language    map[string]string
}

type page struct {
	number    int
	userInput string
}

func newCache() *cache {
	return &cache{
		mu:          sync.Mutex{},
		subscribers: make(map[string][]page),
		language:    make(map[string]string),
	}
}

func (c *cache) checkLanguage(username string) (bool, string) {
	c.mu.Lock()
	l, ok := c.language[username]
	c.mu.Unlock()
	return ok, l
}

func (c *cache) setupLanguage(username, language string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.language[username] = language
	return language
}

func (c *cache) set(k int64, p page) []page {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := keyGen(k)
	val, ok := c.subscribers[key]
	if !ok {
		c.subscribers[key] = make([]page, 0, 4)
	}
	val = append(val, p)
	c.subscribers[key] = val
	return val
}

func keyGen(k int64) string {
	return strconv.FormatInt(k, 10)
}

func (c *cache) get(k int64) (ps []page, ok bool) {
	key := keyGen(k)
	if ps, ok = c.subscribers[key]; !ok {
		return nil, false
	}
	return
}

func (c *cache) back(k int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := keyGen(k)
	ps := c.subscribers[key]
	if len(ps) <= 1 {
		delete(c.subscribers, key)
		return
	}
	c.subscribers[key] = ps[:len(ps)-1]
}

func (c *cache) delete(k int64) {
	c.mu.Lock()
	key := keyGen(k)
	delete(c.subscribers, key)
	c.mu.Unlock()
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

func (p *page) giveContent(page int, language string) (c string) {
	switch language {
	case "russian":
		switch page {
		case 0:
			c = firstPageRUS
		case 1:
			c = secondPageRUS
		case 2:
			c = fmt.Sprintf(thirdPageRUS, strings.ToUpper(p.userInput))
		case 3:
			c = fourthPageRUS
		case -1:
			c = helpMessageRUS
		}
	case "english":
		switch page {
		case 0:
			c = firstPageENG
		case 1:
			c = secondPageENG
		case 2:
			c = fmt.Sprintf(thirdPageENG, strings.ToUpper(p.userInput))
		case 3:
			c = fourthPageENG
		case -1:
			c = helpMessageENG
		}
	}

	return
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
				b.api.Send(msg)
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
				b.api.Send(msg)
				continue
			case "help":
				b.api.Send(help(chatID, language))
				continue
			case "language":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, selectLanguageMsg(language))
				msg.ReplyMarkup = languagesKeyBoard
				b.api.Send(msg)
				continue
			}

			pages, ok := b.cache.get(chatID)
			if !ok {
				b.api.Send(help(chatID, language))
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
				b.cache.back(chatID)
				b.api.Send(msg)
				continue
			}

			if pages[len(pages)-1].number == 0 {
				_, ok := cryptoCurrencies[strings.ToUpper(userText)]
				if !ok {
					msg := tgbotapi.NewMessage(chatID, handleErrorInput(0, language))
					b.api.Send(msg)
					continue
				}
			}

			if pages[len(pages)-1].number == 1 {
				_, ok := fiats[strings.ToUpper(userText)]
				if !ok {
					msg := tgbotapi.NewMessage(chatID, handleErrorInput(1, language))
					b.api.Send(msg)
					continue
				}
			}

			if pages[len(pages)-1].number == 2 {
				_, err := strconv.ParseFloat(userText, 10)
				if err != nil {
					b.api.Send(tgbotapi.NewMessage(chatID, handleErrorInput(2, language)))
					continue
				}
			}

			if pages[len(pages)-1].number == 3 {
				if _, ok := conditionsVarifier[userText]; !ok {
					b.api.Send(tgbotapi.NewMessage(chatID, handleErrorInput(3, language)))
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
				alert := splitArgs(pages[1:], chatID)
				var text string
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
			b.api.Send(msg)
		}
	}
}

func help(id int64, language string) (m tgbotapi.MessageConfig) {
	switch language {
	case "english":
		m = tgbotapi.NewMessage(id, helpMessageENG)
	case "russian":
		m = tgbotapi.NewMessage(id, helpMessageRUS)
	}
	return
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

func backKeyboard(language string) tgbotapi.ReplyKeyboardMarkup {
	var data string
	switch language {
	case "russian":
		data = "назад"
	case "english":
		data = "back"
	}
	return tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(data)))
}

func conditionsKeyboard(language string) tgbotapi.ReplyKeyboardMarkup {
	var backButton []tgbotapi.KeyboardButton
	switch language {
	case "russian":
		backButton = tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("назад"),
		)
	case "english":
		backButton = tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("back"),
		)
	}
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(">="),
			tgbotapi.NewKeyboardButton("=="),
			tgbotapi.NewKeyboardButton("<=")),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(">"),
			tgbotapi.NewKeyboardButton("<"),
		),
		backButton,
	)
}

var languagesKeyBoard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🇷🇺", "russian"),
		tgbotapi.NewInlineKeyboardButtonData("🇺🇸", "english"),
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
