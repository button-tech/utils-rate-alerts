package respond

import (
	"encoding/json"
	"log"

	routing "github.com/qiangxue/fasthttp-routing"
)

func WithJSON(ctx *routing.Context, code int, payload map[string]interface{}) {
	ctx.SetContentType("application/json")
	ctx.SetStatusCode(code)
	if err := json.NewEncoder(ctx).Encode(payload); err != nil {
		log.Println("write answer", err)
	}
}
