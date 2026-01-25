# Hytale News RSS Feed

A lightweight Go application that scrapes the [Hytale news page](https://hytale.com/news) and generates an RSS feed, allowing you to stay updated with the latest Hytale news through your favorite RSS reader.

This was created using Claude AI.

## Features

- 🔄 Automatically scrapes Hytale news every hour
- 📰 Generates RSS 2.0 compatible feed
- 🚀 Lightweight web server using Chi router
- 🐳 Docker support for easy deployment
- 🔒 Thread-safe feed updates

## Disclaimer

This project is **not affiliated with, endorsed by, or connected to Hytale or Hypixel Studios** in any way. This is an unofficial, community-created tool for personal use. All content scraped from hytale.com remains the property of its respective owners.

## Quick Start

### Using Go

```bash
# Clone the repository
git clone <your-repo-url>
cd hytale-rss

# Install dependencies
go mod download

# Run the application
go run main.go
```

The server will start on `http://localhost:8080`

### Using Docker

```bash
# Build the image
docker build -t hytale-rss .

# Run the container
docker run -p 8080:8080 hytale-rss
```

## Usage

Once the server is running, you can access:

- **RSS Feed**: `http://localhost:8080/feed.xml`

Add the feed URL to your RSS reader to receive automatic updates.

## Configuration

The application updates the feed every hour by default. To change this, modify the ticker duration in `main.go`:

```go
ticker := time.NewTicker(1 * time.Hour) // Change to your preferred interval
```

You can also change the server port by modifying:

```go
log.Fatal(http.ListenAndServe(":8080", r)) // Change :8080 to your preferred port
```

## License

MIT License - feel free to use and modify as needed.
