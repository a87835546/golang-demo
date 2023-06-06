package handler

import (
	"context"
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/kataras/iris/v12"
	"golang-demo/internal/consts"
)

type ContextExample struct {
}

func Test(ctx iris.Context) {
	ctx.Values().Get("a")
	context.Background()
	context.TODO()
}
func Test8(ctx iris.Context) {
	c := colly.NewCollector(
		colly.MaxDepth(2),
		colly.AllowURLRevisit(),
	)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Println(e.Text, link)
		// Visit link found on page
		e.Request.Visit(link)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err.Error())
	})
	// Start scraping on https://en.wikipedia.org
	c.Visit("https://github.com/")
	Re(ctx, consts.Success, nil)
}
