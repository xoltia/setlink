package visitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Visitor struct {
	httpClient http.Client
}

func NewVisitor(timeout time.Duration) *Visitor {
	return &Visitor{
		httpClient: http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Visitor) Visit(ctx context.Context, url *url.URL) (*VisitResult, error) {
	log.Printf("Crawling %s\n", url.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	req.Header.Set("User-Agent", fmt.Sprintf("SetLinkCrawler/%s", os.Getenv("VERSION")))
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	result := NewVisitResult(url)
	tkn := html.NewTokenizer(resp.Body)
	isHead := false

	for {
		tt := tkn.Next()
		switch tt {
		case html.ErrorToken:
			return nil, tkn.Err()
		case html.StartTagToken:
			t := tkn.Token()
			if !isHead && t.Data == "head" {
				isHead = true
			}
			if isHead && t.Data == "meta" {
				c.setPotentialCardAttribute(result, t)
			}
		case html.SelfClosingTagToken:
			t := tkn.Token()
			if isHead && t.Data == "meta" {
				c.setPotentialCardAttribute(result, t)
			}
		case html.EndTagToken:
			t := tkn.Token()
			isHead = !(t.Data == "head")
			if !isHead {
				return result, nil
			}
		}
	}
}

func (c *Visitor) setPotentialCardAttribute(result *VisitResult, t html.Token) {
	if len(t.Attr) < 2 {
		return
	}

	propertyParts := strings.Split(t.Attr[0].Val, ":")
	if len(propertyParts) != 2 {
		return
	}

	fmt.Println(propertyParts)

	cardType := propertyParts[0]
	property := propertyParts[1]

	if cardType != "og" && cardType != "twitter" {
		return
	}

	var card *VisitCardResult

	switch cardType {
	case "og":
		card = &result.OpenGraph
	case "twitter":
		card = &result.Twitter
	}

	switch property {
	case "title":
		card.Title = t.Attr[1].Val
	case "description":
		card.Description = t.Attr[1].Val
	case "image":
		for _, attr := range t.Attr {
			if attr.Key == "content" {
				card.Image = attr.Val
				break
			}
		}
	}
}
