package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	processCache "github.com/jeyldii/rate-alerts/pkg/storage/cache"
	"github.com/streadway/amqp"
)

const (
	helpMessageENG        = `/alert - to sign up for a currency notice`
	helpMessageRUS        = `/alert - подписаться на уведомление по валюте`
	languageMessageRUS    = "Выберите язык"
	languageMessageENG    = "Select language"
	alertResultMessageRUS = `ℹ️ Уведомление о %s
Текущая цена: %s %s
Условие выполнения: %s %s %s`
	alertResultMessageENG = `ℹ️ Notification of %s
Current price: %s %s
Execution Condition: %s %s %s`
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
	errCryptoInputRUS     = "❌ Попробуйте другую крипто валюту\nПример: BTC"
	errFiatInputRUS       = "❌ Попробуйте другую фиатную валюту\nПример: USD"
	errPriceInputRUS      = "❌ Введите валидную сумму\nПример: 7000"
	errConditionInputRUS  = "❌ Введите доступное условие\nПример: <= или >= или == или < или >"
	errAlertMsgRUS        = `❌ Произошла ошибка. Попробуйте позже`
	alertMessageRUS       = `✅ Вы подписаны на уведомление`
	noAlertsMessageRUS    = `💤 Вы не подписаны на уведомления`
	invalidAlertNumberRUS = "❌ Неверный номер уведомления"

	errCryptoInputENG     = "❌ Try another crypto currency\nExample: BTC"
	errFiatInputENG       = "❌ Try another fiat currency\nExample: USD"
	errPriceInputENG      = "❌ Enter valid amount\nExample: 7000"
	errConditionInputENG  = "❌ Enter an available condition\nExample: <= или >= или == или < или >"
	errAlertMsgENG        = `❌ An error has occurred. try late`
	alertMessageENG       = `✅ You subscribed to the notification`
	noAlertsMessageENG    = `💤 You have't got alerts`
	invalidAlertNumberENG = "❌ Invalid alert number"
)

func selectAlertNumber(language string) (m string) {
	switch language {
	case "russian":
		m = invalidAlertNumberRUS
	case "english":
		m = invalidAlertNumberENG
	}
	return
}

func selectNoAlertMessage(language string) (m string) {
	switch language {
	case "russian":
		m = noAlertsMessageRUS
	case "english":
		m = noAlertsMessageENG
	}
	return
}

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

var fiats = map[string]struct{}{"AED": {}, "ALL": {}, "AMD": {}, "AOA": {}, "ARS": {}, "AUD": {}, "BAM": {}, "BDT": {}, "BGN": {}, "BHD": {}, "BIF": {}, "BND": {}, "BOB": {}, "BRL": {}, "BSD": {}, "BTC": {}, "BTN": {}, "BWP": {}, "BYN": {}, "CAD": {}, "CDF": {}, "CHF": {}, "CLP": {}, "CNY": {}, "COP": {}, "CRC": {}, "CZK": {}, "DKK": {}, "DOP": {}, "DZD": {}, "EGP": {}, "ETB": {}, "EUR": {}, "GBP": {}, "GEL": {}, "GGP": {}, "GHS": {}, "GIP": {}, "GTQ": {}, "HKD": {}, "HNL": {}, "HRK": {}, "HUF": {}, "IDR": {}, "ILS": {}, "INR": {}, "IQD": {}, "IRR": {}, "ISK": {}, "JMD": {}, "JOD": {}, "JPY": {}, "KES": {}, "KGS": {}, "KHR": {}, "KRW": {}, "KWD": {}, "KZT": {}, "LBP": {}, "LKR": {}, "LSL": {}, "MAD": {}, "MDL": {}, "MMK": {}, "MOP": {}, "MUR": {}, "MWK": {}, "MXN": {}, "MYR": {}, "NAD": {}, "NGN": {}, "NIO": {}, "NOK": {}, "NPR": {}, "NZD": {}, "OMR": {}, "PAB": {}, "PEN": {}, "PGK": {}, "PHP": {}, "PKR": {}, "PLN": {}, "PYG": {}, "QAR": {}, "RON": {}, "RUB": {}, "RWF": {}, "SAR": {}, "SBD": {}, "SEK": {}, "SGD": {}, "SHP": {}, "SZL": {}, "THB": {}, "TMT": {}, "TND": {}, "TOP": {}, "TRY": {}, "TTD": {}, "TWD": {}, "TZS": {}, "UAH": {}, "UGX": {}, "USD": {}, "UYU": {}, "UZS": {}, "VEF": {}, "VND": {}, "VUV": {}, "XAF": {}, "XAU": {}, "XCD": {}, "XOF": {}, "ZAR": {}, "ZMW": {}}
var cryptoCurrencies = map[string]struct{}{"ADA": {}, "AE": {}, "ALGO": {}, "ARDR": {}, "ATOM": {}, "BCD": {}, "BCH": {}, "BCN": {}, "BNB": {}, "BSV": {}, "BTC": {}, "BTG": {}, "BTM": {}, "BTS": {}, "BTT": {}, "CENNZ": {}, "DASH": {}, "DCR": {}, "DGB": {}, "DOGE": {}, "EOS": {}, "ETC": {}, "ETH": {}, "ICX": {}, "IOST": {}, "KMD": {}, "LSK": {}, "LTC": {}, "LUNA": {}, "MONA": {}, "NANO": {}, "NEO": {}, "NRG": {}, "ONT": {}, "QTUM": {}, "RVN": {}, "STEEM": {}, "STRAT": {}, "THETA": {}, "TOMO": {}, "TRX": {}, "VET": {}, "VSYS": {}, "WAVES": {}, "XEM": {}, "XLM": {}, "XMR": {}, "XRP": {}, "XTZ": {}, "XVG": {}, "ZEC": {}, "ZEN": {}, "ZIL": {}}

var conditionsVerifier = map[string]struct{}{
	">":  {},
	"<":  {},
	"==": {},
	">=": {},
	"<=": {},
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

	channel *amqp.Channel
	queue   amqp.Queue
}

type cache struct {
	mu          sync.Mutex
	subscribers map[string][]page
	language    map[string]string
	alerts      map[string][]string
}

type userAlert struct {
	currency  string
	condition string
	price     string
	fiat      string
	lastAlert int
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
		alerts:      make(map[string][]string),
	}
}

func (c *cache) setAlert(chatID int64, currency, fiat, price, condition string) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	val, ok := c.alerts[k]
	if !ok {
		c.alerts[k] = make([]string, 0)
	}
	val = append(val, genAlertValue(currency, fiat, price, condition))
	c.alerts[k] = val
	c.mu.Unlock()
}

func genAlertValue(currency, fiat, price, condition string) string {
	return fmt.Sprintf("%s_%s_%s_%s", currency, fiat, price, condition)
}

func (c *cache) setRawAlerts(chatID int64, alerts []string) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	c.alerts[k] = alerts
	c.mu.Unlock()
	return
}

func (c *cache) getRawAlerts(chatID int64) ([]string, bool) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return nil, ok
	}
	return val, true
}

func (c *cache) getAlerts(chatID int64) (a string, ok bool) {
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return "", ok
	}

	for i, v := range val {
		alertItems := strings.Split(v, "_")
		u := userAlert{
			currency:  alertItems[0],
			fiat:      alertItems[1],
			price:     alertItems[2],
			condition: alertItems[3],
		}
		one := fmt.Sprintf(
			"%s %s %s %s",
			strings.ToUpper(u.currency),
			u.condition,
			u.price,
			strings.ToUpper(u.fiat),
		)
		a += fmt.Sprintf("№%s %s \n", strconv.Itoa(i+1), one)
	}
	return
}

func (c *cache) deleteAlert(chatID int64, alert string) (string, bool) {
	var deleted string
	k := keyGenForAlert(chatID)
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.alerts[k]
	if !ok {
		return "", true
	}

	numb, err := strconv.Atoi(alert)
	if err != nil {
		return "", false
	}

	//if len(val) - numb == 0 {
	//	delete(c.alerts, k)
	//	return "", true
	//}

	if len(val) >= numb {
		deleted = val[numb-1]
		val = append(val[:numb-1], val[numb-1+1:]...)
		c.alerts[k] = val
		return deleted, true
	}

	return "", false
}

func keyGenForAlert(chatID int64) string {
	convChatID := strconv.FormatInt(chatID, 10)
	return fmt.Sprintf("%s_%s", convChatID, "alerts")
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
		api:                 bot,
		tgChannel:           updates,
		channel:             p.Channel,
		queue:               p.Queue,
		cache:               newCache(),
		deleteProcessingURL: os.Getenv("PROCESSING_API_URL"),
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
	return requestToDelete(block, b.deleteProcessingURL)
}

func requestToDelete(b processCache.ConditionBlock, url string) error {
	rq := req.New()
	resp, err := rq.Post(url+"delete", req.BodyJSON(&b))
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

func help(id int64, language string) tgbotapi.MessageConfig {
	var text string
	switch language {
	case "english":
		text = helpMessageENG
	case "russian":
		text = helpMessageRUS
	}
	return tgbotapi.NewMessage(id, text)
}

func completedAlertMessage(c trueCondition, language string) string {
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

func (b *Bot) AlertUser(c trueCondition) error {
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

func splitArgs(args []page, chatID int64, language string) alert {
	convChatID := strconv.FormatInt(chatID, 10)
	l := fmt.Sprintf("%s_%s", convChatID, language)
	return alert{
		Currency:  args[0].userInput,
		Fiat:      args[1].userInput,
		Price:     args[2].userInput,
		Condition: args[3].userInput,
		URL:       l,
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
	var text string
	switch language {
	case "russian":
		text = "назад"
	case "english":
		text = "back"
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
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(text),
		),
	)
}

var languagesKeyBoard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🇷🇺", "russian"),
		tgbotapi.NewInlineKeyboardButtonData("🇺🇸", "english"),
	),
)
