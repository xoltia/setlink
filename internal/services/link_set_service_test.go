package services

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xoltia/setlink/internal/models"
	"github.com/xoltia/setlink/pkg/sliceutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLinkSetServiceCreate(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.LinkSet{}, &models.Link{})

	linkSetService := NewLinkSetService(db)

	sharedServer := makeTestServer()

	sharedUrl := sharedServer.URL

	serverSet1 := []*httptest.Server{
		makeTestServer(),
		makeTestServer(),
		sharedServer,
	}

	serverSet2 := []*httptest.Server{
		makeTestServer(),
		makeTestServer(),
		sharedServer,
	}

	urls := sliceutil.Map(serverSet1, func(s *httptest.Server) *url.URL {
		url, _ := url.Parse(s.URL)
		return url
	})

	ctx := context.Background()
	linkSet, err := linkSetService.GetOrCreateSet(ctx, urls)
	assert.Nil(t, err)

	duplicateLinkSet, err := linkSetService.GetOrCreateSet(ctx, urls)
	assert.Nil(t, err)

	log.Printf("%+v", linkSet)
	log.Printf("%+v", duplicateLinkSet)

	urls2 := sliceutil.Map(serverSet2, func(s *httptest.Server) *url.URL {
		url, _ := url.Parse(s.URL)
		return url
	})

	differentLinkSet, err := linkSetService.GetOrCreateSet(ctx, urls2)
	assert.Nil(t, err)

	log.Printf("%+v", differentLinkSet)

	assert.Equal(t, len(linkSet.Links), len(duplicateLinkSet.Links))
	assert.Equal(t, linkSet.HashString, duplicateLinkSet.HashString)
	assert.Equal(t, linkSet.ID, duplicateLinkSet.ID)

	assert.NotEqual(t, linkSet.HashString, differentLinkSet.HashString)

	sharedLink1 := sliceutil.Find(linkSet.Links, func(l *models.Link) bool {
		return l.URL == sharedUrl
	})

	sharedLink2 := sliceutil.Find(differentLinkSet.Links, func(l *models.Link) bool {
		return l.URL == sharedUrl
	})

	assert.NotNil(t, sharedLink1)
	assert.NotNil(t, sharedLink2)

	assert.Equal(t, (*sharedLink1).HashString, (*sharedLink2).HashString)
	assert.Equal(t, (*sharedLink1).ID, (*sharedLink2).ID)
}

func makeTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>
			<head>
				<title>Test</title>
			    <meta name="og:title" content="Test">
			    <meta name="og:description" content="Test description">
				<meta name="og:image" content="https://example.com/image.png">
				<meta name="twitter:title" content="Test">
				<meta name="twitter:description" content="Test description">
				<meta name="twitter:image" content="https://example.com/image.png">
			</head>
		</html>`))
	}))
}
