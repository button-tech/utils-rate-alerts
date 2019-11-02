package receiver

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/button-tech/rate-alerts/storage"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type requestBlocks struct {
	Tokens     []string `json:"tokens"`
	Currencies []string `json:"currencies"`
	API        string   `json:"api"`
}

func (r *Receiver) deliveryChannel() (<-chan amqp.Delivery, error) {
	msgs, err := r.rabbitMQ.Channel.Consume(
		r.rabbitMQ.Queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (r *Receiver) Processing() {
	c, err := r.deliveryChannel()
	if err != nil {
		log.Println(err)
		return
	}

	for msg := range c {
		var block storage.ConditionBlock
		if err := json.Unmarshal(msg.Body, &block); err != nil {
			log.Println(err)
			continue
		}
		r.store.Set(block)
	}

	select {}
}

const cmc = "cmc"

func (r *Receiver) GetPrices() {
	t := time.NewTicker(time.Second * 30)
	for ; ; <-t.C {
		var blocks requestBlocks
		blocks.API = cmc

		for k := range r.checkMap(&blocks) {
			blocks.Currencies = append(blocks.Currencies, k)
		}

		if err := r.getPrices(&blocks); err != nil {
			log.Println(err)
		}
	}
}

func (r *Receiver) checkMap(blocks *requestBlocks) map[string]struct{} {
	stored := r.store.Get()

	m := make(map[string]struct{})
	for currency, fiat := range stored {
		blocks.Tokens = append(blocks.Tokens, string(currency))
		for f := range fiat {
			if _, ok := m[string(f)]; !ok {
				m = map[string]struct{}{}
			}
			m[string(f)] = struct{}{}
		}
	}
	return m
}

func (r *Receiver) getPrices(b *requestBlocks) error {
	gotPrices, err := doRequest(b)
	if err != nil {
		return err
	}

	if err := r.schedule(gotPrices); err != nil {
		return err
	}

	return nil
}

func doRequest(b *requestBlocks) ([]*parsedPrices, error) {
	rq := req.New()
	resp, err := rq.Post(os.Getenv("PRICES"), req.BodyJSON(&b))
	if err != nil {
		return nil, err
	}
	if resp.Response().StatusCode != fasthttp.StatusOK {
		return nil, errors.Wrap(errors.New("No http statusOK"), "responseStatusCode")
	}

	ps, err := respFastJSON(resp.Bytes())
	if err != nil {
		return nil, err
	}

	return ps, nil
}

type parsedPrices struct {
	currency string
	rates    map[string]string
}

const (
	currency = "currency"
	rates    = "rates"
)

func respFastJSON(b []byte) ([]*parsedPrices, error) {
	var p fastjson.Parser
	parsed, err := p.ParseBytes(b)
	if err != nil {
		return nil, errors.Wrap(err, "parseBytes")
	}

	var pp []*parsedPrices

	o := parsed.GetObject()
	data := o.Get("data")
	array, err := data.Array()
	if err != nil {
		return nil, errors.Wrap(err, "can't get array")
	}
	for _, v := range array {
		obj, err := v.Object()
		if err != nil {
			return nil, errors.Wrap(err, "can't get object")
		}

		var p parsedPrices
		m := make(map[string]string)
		obj.Visit(func(key []byte, v *fastjson.Value) {
			sKey := string(key)
			if sKey == currency {
				p.currency = trim(v.String())
			}

			if sKey == rates {
				rates, _ := v.Array()
				for _, rate := range rates {
					rateObj, _ := rate.Object()
					rateObj.Visit(func(key []byte, v *fastjson.Value) {
						m[string(key)] = trim(v.String())
					})
				}
				p.rates = m
			}
		})
		pp = append(pp, &p)
	}

	return pp, nil
}

func trim(s string) string {
	trimmed := strings.TrimPrefix(s, "\"")
	trimmed = strings.TrimSuffix(trimmed, "\"")
	return trimmed
}

func (r *Receiver) schedule(pp []*parsedPrices) error {
	stored := r.store.Get()
	var requests []storage.ConditionBlock

	for _, p := range pp {
		for token, price := range p.rates {
			urlMap := stored[storage.Token(token)][storage.Fiat(p.currency)]
			for _, block := range urlMap {
				parsedFloats, err := parseFloat(price, block.Price)
				if err != nil {
					return err
				}

				currentPrice := parsedFloats[0]
				conditionPrice := parsedFloats[1]
				if block.Condition == "==" && currentPrice == conditionPrice ||
					block.Condition == ">=" && currentPrice >= conditionPrice ||
					block.Condition == "<=" && currentPrice <= conditionPrice {
					requests = append(requests, block)
				}
			}
		}
	}

	if len(requests) == 0 {
		return errors.New("no block to process")
	}

	for _, block := range requests {
		go r.checkStatusAccepted(block)
	}
	return nil
}

func parseFloat(f, s string) ([]float64, error) {
	var floats []float64
	first, err := strconv.ParseFloat(f, 64)
	if err != nil {
		return nil, err
	}

	second, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, err
	}

	floats = append(floats, first, second)
	return floats, nil
}

func (r *Receiver) checkStatusAccepted(block storage.ConditionBlock) {
	var err error
	t := time.NewTicker(time.Second * 3)

	counter := 0
	for ; counter < 4; <-t.C {
		err = checkURL(block.URL)
		if err == nil {
			if err := r.store.Delete(block); err != nil {
				log.Println(err)
				continue
			}
			break
		}
		counter++
	}

	if err != nil {
		log.Println(err)
	}
}

func checkURL(url string) error {
	rq := req.New()
	_, err := rq.Get(url)
	if err != nil {
		return errors.Wrap(err, "checkURL")
	}

	// Comment for test
	//if resp.Response().StatusCode != 202 {
	//	return errors.Wrap(errors.New("response statusCode not 202"), "checkURL")
	//}

	return nil
}
