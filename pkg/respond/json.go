package respond

import (
	"encoding/json"
	"log"

	t "github.com/button-tech/utils-rate-alerts/types"
	routing "github.com/qiangxue/fasthttp-routing"
)

func WithJSON(ctx *routing.Context, code int, payload t.Payload) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(payload); err != nil {
		log.Println("write answer", err)
	}
}
