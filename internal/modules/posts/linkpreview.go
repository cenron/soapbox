package posts

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

var urlRegex = regexp.MustCompile(`https?://[^\s<>"]+`)

const linkPreviewTimeout = 5 * time.Second

type linkPreviewData struct {
	URL         string
	Title       string
	Description string
	ImageURL    string
}

func extractFirstURL(body string) string {
	match := urlRegex.FindString(body)
	return match
}

func fetchLinkPreview(ctx context.Context, rawURL string) *linkPreviewData {
	ctx, cancel := context.WithTimeout(ctx, linkPreviewTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "SoapboxBot/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}

	return parseHTMLMeta(rawURL, string(body))
}

func parseHTMLMeta(rawURL, htmlBody string) *linkPreviewData {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil
	}

	data := &linkPreviewData{URL: rawURL}
	extractMeta(doc, data)

	if data.Title == "" {
		data.Title = extractTitle(doc)
	}

	if data.Title == "" && data.Description == "" {
		return nil
	}

	return data
}

func extractMeta(n *html.Node, data *linkPreviewData) {
	if n.Type == html.ElementNode && n.Data == "meta" {
		var property, name, content string
		for _, attr := range n.Attr {
			switch attr.Key {
			case "property":
				property = attr.Val
			case "name":
				name = attr.Val
			case "content":
				content = attr.Val
			}
		}

		switch {
		case property == "og:title" && data.Title == "":
			data.Title = content
		case property == "og:description" && data.Description == "":
			data.Description = content
		case property == "og:image" && data.ImageURL == "":
			data.ImageURL = content
		case name == "description" && data.Description == "":
			data.Description = content
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractMeta(c, data)
	}
}

func extractTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" {
		if n.FirstChild != nil {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if t := extractTitle(c); t != "" {
			return t
		}
	}
	return ""
}
