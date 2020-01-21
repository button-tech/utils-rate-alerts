package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
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

var fiats = map[string]struct{}{"AED": {}, "ALL": {}, "AMD": {}, "AOA": {}, "ARS": {}, "AUD": {}, "BAM": {}, "BDT": {}, "BGN": {}, "BHD": {}, "BIF": {}, "BND": {}, "BOB": {}, "BRL": {}, "BSD": {}, "BTC": {}, "BTN": {}, "BWP": {}, "BYN": {}, "CAD": {}, "CDF": {}, "CHF": {}, "CLP": {}, "CNY": {}, "COP": {}, "CRC": {}, "CZK": {}, "DKK": {}, "DOP": {}, "DZD": {}, "EGP": {}, "ETB": {}, "EUR": {}, "GBP": {}, "GEL": {}, "GGP": {}, "GHS": {}, "GIP": {}, "GTQ": {}, "HKD": {}, "HNL": {}, "HRK": {}, "HUF": {}, "IDR": {}, "ILS": {}, "INR": {}, "IQD": {}, "IRR": {}, "ISK": {}, "JMD": {}, "JOD": {}, "JPY": {}, "KES": {}, "KGS": {}, "KHR": {}, "KRW": {}, "KWD": {}, "KZT": {}, "LBP": {}, "LKR": {}, "LSL": {}, "MAD": {}, "MDL": {}, "MMK": {}, "MOP": {}, "MUR": {}, "MWK": {}, "MXN": {}, "MYR": {}, "NAD": {}, "NGN": {}, "NIO": {}, "NOK": {}, "NPR": {}, "NZD": {}, "OMR": {}, "PAB": {}, "PEN": {}, "PGK": {}, "PHP": {}, "PKR": {}, "PLN": {}, "PYG": {}, "QAR": {}, "RON": {}, "RUB": {}, "RWF": {}, "SAR": {}, "SBD": {}, "SEK": {}, "SGD": {}, "SHP": {}, "SZL": {}, "THB": {}, "TMT": {}, "TND": {}, "TOP": {}, "TRY": {}, "TTD": {}, "TWD": {}, "TZS": {}, "UAH": {}, "UGX": {}, "USD": {}, "UYU": {}, "UZS": {}, "VEF": {}, "VND": {}, "VUV": {}, "XAF": {}, "XAU": {}, "XCD": {}, "XOF": {}, "ZAR": {}, "ZMW": {}}
var cryptoCurrencies = map[string]struct{}{"ADA": {}, "AE": {}, "ALGO": {}, "ARDR": {}, "ATOM": {}, "BCD": {}, "BCH": {}, "BCN": {}, "BNB": {}, "BSV": {}, "BTC": {}, "BTG": {}, "BTM": {}, "BTS": {}, "BTT": {}, "CENNZ": {}, "DASH": {}, "DCR": {}, "DGB": {}, "DOGE": {}, "EOS": {}, "ETC": {}, "ETH": {}, "ICX": {}, "IOST": {}, "KMD": {}, "LSK": {}, "LTC": {}, "LUNA": {}, "MONA": {}, "NANO": {}, "NEO": {}, "NRG": {}, "ONT": {}, "QTUM": {}, "RVN": {}, "STEEM": {}, "STRAT": {}, "THETA": {}, "TOMO": {}, "TRX": {}, "VET": {}, "VSYS": {}, "WAVES": {}, "XEM": {}, "XLM": {}, "XMR": {}, "XRP": {}, "XTZ": {}, "XVG": {}, "ZEC": {}, "ZEN": {}, "ZIL": {}}

var conditionsVerifier = map[string]struct{}{
	">":  {},
	"<":  {},
	"==": {},
	">=": {},
	"<=": {},
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
