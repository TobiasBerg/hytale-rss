package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/net/html"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type NewsItem struct {
	Title string
	Link  string
	Date  string
	Desc  string
}

var (
	feedMutex sync.RWMutex
	rssFeed   *RSS
)

func main() {
	// Initial scrape
	if err := updateFeed(); err != nil {
		log.Printf("Initial scrape failed: %v", err)
	}

	// Start background updater (every hour)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := updateFeed(); err != nil {
				log.Printf("Failed to update feed: %v", err)
			} else {
				log.Println("Feed updated successfully")
			}
		}
	}()

	// HTTP server with Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Routes
	r.Get("/", homeHandler)
	r.Get("/feed.xml", serveFeed)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func updateFeed() error {
	log.Println("Scraping Hytale news...")

	resp, err := http.Get("https://hytale.com/news")
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	items := scrapeNewsItems(doc)

	feed := &RSS{
		Version: "2.0",
		Channel: Channel{
			Title:       "Hytale News",
			Link:        "https://hytale.com/news",
			Description: "Latest news from Hytale",
			Items:       make([]Item, 0, len(items)),
		},
	}

	for _, item := range items {
		feed.Channel.Items = append(feed.Channel.Items, Item{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Desc,
			PubDate:     item.Date,
		})
	}

	feedMutex.Lock()
	rssFeed = feed
	feedMutex.Unlock()

	log.Printf("Scraped %d news items", len(items))
	return nil
}

func scrapeNewsItems(n *html.Node) []NewsItem {
	var items []NewsItem
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" && hasClass(n, "postWrapper") {
			item := extractPost(n)
			if item.Title != "" {
				items = append(items, item)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return items
}

func extractPost(n *html.Node) NewsItem {
	item := NewsItem{}

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h4":
				if hasClass(n, "post__details__heading") {
					item.Title = strings.TrimSpace(getTextContent(n))
				}
			case "a":
				if hasClass(n, "post") {
					if href := getAttr(n, "href"); href != "" {
						if !strings.HasPrefix(href, "http") {
							item.Link = "https://hytale.com" + href
						} else {
							item.Link = href
						}
					}
				}
			case "span":
				if hasClass(n, "post__details__meta") {
					item.Date = getDateAttr(n)
				}
				if item.Desc == "" && hasClass(n, "post__details__body") {
					desc := getTextContent(n)
					// Truncate description to ~200 chars
					if len(desc) > 200 {
						desc = desc[:200] + "..."
					}
					item.Desc = strings.TrimSpace(desc)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	extract(n)

	return item
}

func getTextContent(n *html.Node) string {
	var text string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			text += n.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return strings.TrimSpace(text)
}

func getDateAttr(n *html.Node) string {
	replacer := strings.NewReplacer("st", "", "nd", "", "rd", "", "th", "")

	date := ""
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "span" {
			if hasClass(c, "post__details__meta__date") {
				fullDate := getTextContent(c)
				if fullDate != "" {
					parsedDate, err := time.Parse("January 2 2006", replacer.Replace(fullDate))
					if err == nil {
						date = parsedDate.Format(time.RFC1123Z)
					}
				}
			}
		}
	}

	return date
}

func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			if slices.Contains(classes, className) {
				return true
			}
		}
	}
	return false
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<html><body><h1>Hytale News RSS Feed</h1><p>Access the feed at: <a href="/feed.xml">/feed.xml</a></p></body></html>`)
}

func serveFeed(w http.ResponseWriter, r *http.Request) {
	feedMutex.RLock()
	feed := rssFeed
	feedMutex.RUnlock()

	if feed == nil {
		http.Error(w, "Feed not ready yet", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		http.Error(w, "Failed to generate feed", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(xml.Header))
	w.Write(output)
}
