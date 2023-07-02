package services

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/xoltia/setlink/internal/models"
	"github.com/xoltia/setlink/internal/visitor"
	"gorm.io/gorm"
)

var ErrNotFound = gorm.ErrRecordNotFound

type LinkSetService struct {
	db      *gorm.DB
	crawler *visitor.Visitor
}

func NewLinkSetService(db *gorm.DB) *LinkSetService {
	return &LinkSetService{
		db:      db,
		crawler: visitor.NewVisitor(10 * time.Second),
	}
}

func (l *LinkSetService) GetByID(ctx context.Context, id int) (*models.LinkSet, error) {
	var linkSet models.LinkSet

	result := l.db.Preload("Links").First(&linkSet, id)

	if result.Error != nil {
		return nil, result.Error
	}

	return &linkSet, nil
}

func (l *LinkSetService) GetByHash(ctx context.Context, hash string) (*models.LinkSet, error) {
	var linkSet models.LinkSet

	result := l.db.Where("hash_string = ?", hash).Preload("Links").First(&linkSet)

	if result.Error != nil {
		return nil, result.Error
	}

	return &linkSet, nil
}

func (l *LinkSetService) GetOrCreateSet(ctx context.Context, urls []*url.URL) (*models.LinkSet, error) {
	var links []*models.Link

	for _, url := range urls {
		link := &models.Link{URL: url.String()}
		links = append(links, link)
	}

	linkSet := &models.LinkSet{Links: links}
	_, err := linkSet.ComputeHash()

	if err != nil {
		return nil, err
	}

	var existingLinkSet models.LinkSet

	log.Println(linkSet.HashString)

	result := l.db.Where("hash_string = ?", linkSet.HashString).Preload("Links").First(&existingLinkSet)

	if result.Error == nil {
		log.Println("Found existing link set")
		return &existingLinkSet, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	results := make(chan *visitor.VisitResult)
	errs := make(chan error)

	for _, u := range urls {
		go func(ctx context.Context, crawler *visitor.Visitor, url *url.URL) {
			result, err := crawler.Visit(ctx, url)
			if err != nil {
				errs <- err
			}
			results <- result
		}(ctx, l.crawler, u)
	}

	newSet := &models.LinkSet{Links: []*models.Link{}}

	for i := 0; i < len(urls); i++ {
		select {
		case result := <-results:
			newLink := crawlResultToLink(result)
			newLink.ComputeHash()
			var existingLink models.Link
			dbResult := l.db.Where("hash_string = ?", newLink.HashString).First(&existingLink)

			if dbResult.Error == nil {
				newSet.Links = append(newSet.Links, &existingLink)
				continue
			}

			if dbResult.Error != gorm.ErrRecordNotFound {
				return nil, dbResult.Error
			}

			l.db.Create(newLink)
			newSet.Links = append(newSet.Links, newLink)
		case err := <-errs:
			return nil, err
		}
	}

	dbResult := l.db.Create(newSet)

	if dbResult.Error != nil {
		return nil, dbResult.Error
	}

	return newSet, nil
}

func crawlResultToLink(crawlResult *visitor.VisitResult) *models.Link {
	link := &models.Link{
		URL:     crawlResult.URL,
		Favicon: crawlResult.Favicon,
	}

	if crawlResult.OpenGraph.Title != "" {
		link.Title = crawlResult.OpenGraph.Title
	} else {
		link.Title = crawlResult.Twitter.Title
	}

	if crawlResult.OpenGraph.Description != "" {
		link.Description = crawlResult.OpenGraph.Description
	} else {
		link.Description = crawlResult.Twitter.Description
	}

	if crawlResult.OpenGraph.Image != "" {
		link.Image = crawlResult.OpenGraph.Image
	} else {
		link.Image = crawlResult.Twitter.Image
	}

	return link
}
