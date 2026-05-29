package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type GitHubRepo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Language    string `json:"language"`
	Archived    bool   `json:"archived"`
	Fork        bool   `json:"fork"`
}

type GitHubStats struct {
	TotalRepos  int
	TotalStars  int
	TotalForks  int
	TopLanguage string
	Followers   int
}

var (
	cache     GitHubStats
	cacheMut  sync.Mutex
	cacheTime time.Time
	cacheTTL  = 5 * time.Minute
)

func bearerReq(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

func fetchJSON(url string, target any, client *http.Client) bool {
	req := bearerReq(url)
	if req == nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// token bad — retry without auth
		io.Copy(io.Discard, resp.Body)
		req2, _ := http.NewRequest("GET", url, nil)
		req2.Header.Set("Accept", "application/vnd.github.v3+json")
		resp2, err2 := client.Do(req2)
		if err2 != nil {
			return false
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			io.Copy(io.Discard, resp2.Body)
			return false
		}
		return json.NewDecoder(resp2.Body).Decode(target) == nil
	}

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return false
	}
	return json.NewDecoder(resp.Body).Decode(target) == nil
}

func FetchGitHubStats(username string) GitHubStats {
	cacheMut.Lock()
	if !cacheTime.IsZero() && time.Since(cacheTime) < cacheTTL {
		cached := cache
		cacheMut.Unlock()
		return cached
	}
	cacheMut.Unlock()

	stats := GitHubStats{}
	langCount := make(map[string]int)
	client := &http.Client{Timeout: 10 * time.Second}

	// profile — get follower count
	var profile struct{ Followers int `json:"followers"` }
	if fetchJSON("https://api.github.com/users/"+username, &profile, client) {
		stats.Followers = profile.Followers
	}

	// repos — paginate all
	fetchedAny := false
	page := 1
	for {
		var repos []GitHubRepo
		url := "https://api.github.com/users/" + username + "/repos?per_page=100&page=" + strconv.Itoa(page)
		if !fetchJSON(url, &repos, client) {
			break
		}
		if len(repos) == 0 {
			break
		}
		fetchedAny = true

		for _, r := range repos {
			if r.Archived || r.Fork {
				continue
			}
			stats.TotalRepos++
			stats.TotalStars += r.Stars
			stats.TotalForks += r.Forks
			if r.Language != "" {
				langCount[r.Language]++
			}
		}

		if len(repos) < 100 {
			break
		}
		page++
	}

	if fetchedAny {
		topLang := ""
		maxCount := 0
		for lang, count := range langCount {
			if count > maxCount {
				maxCount = count
				topLang = lang
			}
		}
		stats.TopLanguage = topLang

		cacheMut.Lock()
		cache = stats
		cacheTime = time.Now()
		cacheMut.Unlock()
	} else {
		// fetch failed — serve stale cache if available
		cacheMut.Lock()
		if !cacheTime.IsZero() {
			stats = cache
		} else {
			log.Println("github: first fetch failed, returning zeros")
		}
		cacheMut.Unlock()
	}

	return stats
}
