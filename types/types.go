package types

type Payload map[string]interface{}

type Alert struct {
	Currency  string `json:"currency"`
	Price     string `json:"price"`
	Fiat      string `json:"fiat"`
	Condition string `json:"condition"`
	URL       string `json:"url"`
}

type TrueCondition struct {
	Result string `json:"result"`
	Values struct {
		Currency     string `json:"currency"`
		Condition    string `json:"condition"`
		Fiat         string `json:"fiat"`
		Price        string `json:"price"`
		CurrentPrice string `json:"currentPrice"`
	} `json:"values"`
	URL string `json:"url"`
}

type RequestBlocks struct {
	Tokens     []string `json:"tokens"`
	Currencies []string `json:"currencies"`
	API        string   `json:"api"`
}
