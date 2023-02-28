package test

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"testing"
)

/**
interface 常见的使用介绍
*/
func init() {
	fmt.Println("init func .....")
	c := colly.NewCollector(
		colly.MaxDepth(1))

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		c.Visit(e.Request.AbsoluteURL(link))

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://baidu.com/")
}

func TestA(t *testing.T) {

}
