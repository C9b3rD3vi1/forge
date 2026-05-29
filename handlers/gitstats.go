package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/html"
)


func GitHubStatsHandler(c *fiber.Ctx) error {
	// Get repository details
	repoResp, err := http.Get("https://api.github.com/repos/C9b3rD3vi1/forge")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "GitHub Repo API error"})
	}
	defer repoResp.Body.Close()

	var repoData map[string]interface{}
	if err := json.NewDecoder(repoResp.Body).Decode(&repoData); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode repo data"})
	}

	// Get contributors
	contribResp, err := http.Get("https://api.github.com/repos/C9b3rD3vi1/forge/contributors")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "GitHub Contributors API error"})
	}
	defer contribResp.Body.Close()

	var contributors []map[string]interface{}
	if err := json.NewDecoder(contribResp.Body).Decode(&contributors); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode contributors"})
	}

	// Prepare response
	return c.JSON(fiber.Map{
		"stars":        repoData["stargazers_count"],
		"forks":        repoData["forks_count"],
		"open_issues":  repoData["open_issues_count"],
		"contributors": len(contributors),
	})
}



func GitHubUserStatsHandler(c *fiber.Ctx) error {
	username := "C9b3rD3vi1"
	log.Printf("Fetching GitHub stats for user: %s", username)

	// --- Fetch user profile ---
	userResp, err := http.Get("https://api.github.com/users/" + username)
	if err != nil {
		log.Printf("ERROR: failed to fetch user profile: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "GitHub User API error"})
	}
	defer userResp.Body.Close()

	var userData map[string]interface{}
	if err := json.NewDecoder(userResp.Body).Decode(&userData); err != nil {
		log.Printf("ERROR: failed to decode user data: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to decode user data"})
	}

	// --- Fetch contributions page ---
	contriResp, err := http.Get("https://github.com/users/" + username + "/contributions")
	if err != nil {
		log.Printf("ERROR: failed to fetch contributions: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "GitHub Contributions API error"})
	}
	defer contriResp.Body.Close()

	bodyBytes, err := io.ReadAll(contriResp.Body)
	if err != nil {
		log.Printf("ERROR: failed to read contributions page: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read contributions page"})
	}

	// --- Parse HTML to find <svg> ---
	doc, err := html.Parse(strings.NewReader(string(bodyBytes)))
	if err != nil {
		log.Printf("ERROR: failed to parse contributions HTML: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse contributions HTML"})
	}

	var svgBuilder strings.Builder
	var findSVG func(*html.Node)
	findSVG = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "svg" {
			// Render this <svg> and stop
			html.Render(&svgBuilder, n)
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			findSVG(child)
		}
	}
	findSVG(doc)

	svg := svgBuilder.String()

	if svg == "" {
		log.Println("⚠️ No <svg> found in contributions page")
	} else {
		log.Printf("Extracted contributions SVG length: %d bytes", len(svg))

		// 🎨 Recolor GitHub greens → Indigo palette
		colorMap := map[string]string{
			"#ebedf0": "#f3f4f6",
			"#9be9a8": "#c7d2fe",
			"#40c463": "#818cf8",
			"#30a14e": "#4f46e5",
			"#216e39": "#312e81",
		}
		for old, new := range colorMap {
			svg = strings.ReplaceAll(svg, old, new)
		}
	}

	// --- Build response ---
	response := fiber.Map{
		"public_repos":      userData["public_repos"],
		"followers":         userData["followers"],
		"following":         userData["following"],
		"gists":             userData["public_gists"],
		"contributions_svg": svg,
	}

	log.Printf("Returning GitHub stats response")
	return c.JSON(response)
}

var (
	chartCache     []byte
	chartCacheMut  sync.Mutex
	chartCacheTime time.Time
	chartCacheTTL  = 15 * time.Minute
)

func ContributionChartHandler(c *fiber.Ctx) error {
	chartCacheMut.Lock()
	if !chartCacheTime.IsZero() && time.Since(chartCacheTime) < chartCacheTTL && len(chartCache) > 0 {
		data := chartCache
		chartCacheMut.Unlock()
		c.Set("Content-Type", "image/svg+xml;charset=utf-8")
		c.Set("Cache-Control", "public, max-age=900")
		return c.Send(data)
	}
	chartCacheMut.Unlock()

	username := "C9b3rD3vi1"
	var dbUser string
	database.DB.Model(&models.Setting{}).Where("key = ?", "github_username").Select("value").Scan(&dbUser)
	if dbUser != "" {
		username = dbUser
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", "https://ghchart.rshah.org/"+username, nil)
	req.Header.Set("User-Agent", "forge/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return c.Status(502).SendString("Failed to fetch contribution chart")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.Status(502).SendString("Contribution chart service returned " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(502).SendString("Failed to read contribution chart")
	}

	svg := string(body)

	// Inject viewBox so the SVG scales with width
	svg = strings.Replace(svg, `<svg version="1.1"`, `<svg version="1.1" viewBox="0 0 663 104"`, 1)

	// Add a dark background rect as the first child of the SVG
	svg = strings.Replace(svg, `x="27" y="20"`, `x="0" y="0" width="663" height="104" fill="#0d1117" rx="4"/><rect x="27" y="20"`, 1)

	// Recolor to GitHub-dark theme for better visibility
	colorMap := map[string]string{
		"#eeeeee": "#161b22",
		"#c6e48b": "#0e4429",
		"#7bc96f": "#006d32",
		"#239a3b": "#26a641",
		"#196127": "#39d353",
	}
	for old, new := range colorMap {
		svg = strings.ReplaceAll(svg, old, new)
	}

	result := []byte(svg)

	chartCacheMut.Lock()
	chartCache = result
	chartCacheTime = time.Now()
	chartCacheMut.Unlock()

	c.Set("Content-Type", "image/svg+xml;charset=utf-8")
	c.Set("Cache-Control", "public, max-age=900")
	return c.Send(result)
}
