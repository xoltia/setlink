package visitor

import (
	"net/url"
)

type VisitCardResult struct {
	Image       string
	Title       string
	Description string
}

type VisitResult struct {
	URL       string
	Favicon   string
	OpenGraph VisitCardResult
	Twitter   VisitCardResult
}

func NewVisitResult(url *url.URL) *VisitResult {
	return &VisitResult{
		OpenGraph: VisitCardResult{},
		Twitter:   VisitCardResult{},
		URL:       url.String(),
		Favicon:   "https://www.google.com/s2/favicons?sz=64&domain=" + url.Host,
	}
}
