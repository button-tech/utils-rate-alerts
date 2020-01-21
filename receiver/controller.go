package receiver

import (
	"encoding/json"
	"github.com/jeyldii/rate-alerts/pkg/respond"
	"github.com/jeyldii/rate-alerts/pkg/storage/cache"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
	"net/http"
)

type controller struct {
	store *cache.Cache
}

func (c *controller) deleteFromProcessing(ctx *routing.Context) error {
	var b cache.ConditionBlock
	if err := json.Unmarshal(ctx.PostBody(), &b); err != nil {
		return err
	}

	if err := c.store.Delete(b); err != nil {
		return err
	}
	respond.WithJSON(ctx, fasthttp.StatusCreated, map[string]interface{}{"result": "ok"})
	return nil
}

func (r *Receiver) mount() {
	r.g.Post("/delete", r.c.deleteFromProcessing)
}

func cors(ctx *routing.Context) error {
	ctx.Response.Header.Set("Access-Control-Allow-Origin", string(ctx.Request.Header.Peek("Origin")))
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", "false")
	ctx.Response.Header.Set("Access-Control-Allow-Methods", "GET,HEAD,PUT,POST,DELETE")
	ctx.Response.Header.Set(
		"Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization",
	)

	if string(ctx.Method()) == "OPTIONS" {
		ctx.Abort()
	}
	if err := ctx.Next(); err != nil {
		if httpError, ok := err.(routing.HTTPError); ok {
			ctx.Response.SetStatusCode(httpError.StatusCode())
		} else {
			ctx.Response.SetStatusCode(http.StatusInternalServerError)
		}

		b, err := json.Marshal(err)
		if err != nil {
			respond.WithJSON(ctx, fasthttp.StatusInternalServerError, map[string]interface{}{"error": err})
			return nil
		}
		ctx.SetContentType("application/json")
		ctx.SetBody(b)
	}
	return nil
}
