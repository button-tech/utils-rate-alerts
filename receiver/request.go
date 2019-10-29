package receiver

type pricesRequest struct {
	Tokens     []string `json:"tokens"`
	Currencies []string `json:"currencies"`
	Change     string   `json:"change"`
	API        string   `json:"api"`
}
