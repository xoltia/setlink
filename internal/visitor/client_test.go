package visitor

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestVisitorClientValidMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>
		<head>
			<title>Test</title>
			<meta property="og:title" content="Test" />
			<meta property="og:description" content="Test description" >
			<meta property="og:image" content="https://example.com/image.png" />
			<meta property="twitter:title" content="Test Twitter" >
			<meta property="twitter:description" content="Test Twitter description" />
			<meta property="twitter:image" content="https://example.com/twitter_image.png" >
		</head>
		<body>
			<h1>Test</h1>
		</body>
		</html>`))
	}))

	client := NewVisitor(5 * time.Second)
	if client == nil {
		t.Error("Expected client to not be nil")
	}

	ctx := context.Background()
	url, _ := url.Parse(server.URL)

	result, err := client.Visit(ctx, url)

	_, faviconErr := url.Parse(result.Favicon)

	assert.Nil(t, err)
	assert.Equal(t, server.URL, result.URL)
	assert.Nil(t, faviconErr)
	assert.NotNil(t, result)
	assert.Equal(t, "https://example.com/image.png", result.OpenGraph.Image)
	assert.Equal(t, "Test", result.OpenGraph.Title)
	assert.Equal(t, "Test description", result.OpenGraph.Description)
	assert.Equal(t, "https://example.com/twitter_image.png", result.Twitter.Image)
	assert.Equal(t, "Test Twitter", result.Twitter.Title)
	assert.Equal(t, "Test Twitter description", result.Twitter.Description)
	t.Cleanup(server.Close)
}

func TestVisitorClientIncomplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>
		<head>
			<title>Test</title>
			<meta property="og:title" content="Test" />
			<meta property="og:description" content="Test description" />
			<meta property="og:image" content="https://example.com/image.png" />
			<meta property="twitter:title" content="Test Twitter" />
			<meta property="twitter:description" content="Test Twitter description" />
			<meta property="twitter:image" content="https://example.com/twitter_image.png" />`))
	}))

	client := NewVisitor(5 * time.Second)
	if client == nil {
		t.Error("Expected client to not be nil")
	}

	ctx := context.Background()
	url, _ := url.Parse(server.URL)

	_, err := client.Visit(ctx, url)

	assert.NotNil(t, err)

	t.Cleanup(server.Close)
}

func TestVisitorClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html>
		<head>
			<title>Test</title>
			<meta property="og:title" content="Test" />
			<meta property="og:description" content="Test description" />
			<meta property="og:image" content="https://example.com/image.png" />
			<meta property="twitter:title" content="Test Twitter" />
			<meta property="twitter:description" content="Test Twitter description" />
			<meta property="twitter:image" content="https://example.com/twitter_image.png" />
		</head>
		<body>
			<h1>Test</h1>
		</body>
		</html>`))
	}))

	client := NewVisitor(10 * time.Millisecond)
	if client == nil {
		t.Error("Expected client to not be nil")
	}

	ctx := context.Background()
	url, _ := url.Parse(server.URL)

	_, err := client.Visit(ctx, url)

	if err, ok := err.(net.Error); !ok || !err.Timeout() {
		t.Error("Expected timeout error")
	}

	t.Cleanup(server.Close)
}
