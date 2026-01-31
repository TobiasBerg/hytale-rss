# Hytale News RSS Feed

The [hytale](https://hytale.com/) site does not have an RSS feed. This project scrapes the Hytale news page and generates an RSS feed that updates automatically every hour.

## Disclaimer

This project is **not affiliated with, endorsed by, or connected to Hytale or Hypixel Studios** in any way. This is an unofficial, community-created tool for personal use. All content scraped from hytale.com remains the property of its respective owners.

## Quick Start

```bash
# Clone the repository
git clone <your-repo-url>
cd hytale-rss

# Install dependencies
go mod download

# Run the application
go run main.go
```

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

## License

MIT License - feel free to use and modify as needed.
