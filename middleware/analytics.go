package middleware

import (
	"strings"
	"sync"
	"time"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

var (
	viewDedup   = make(map[string]time.Time)
	viewDedupMu sync.Mutex
	dedupTTL    = 5 * time.Minute
)

func trackPageView(c *fiber.Ctx, entity, entityID string) {
	path := c.Path()
	ip := c.IP()

	key := ip + ":" + path
	viewDedupMu.Lock()
	if last, ok := viewDedup[key]; ok && time.Since(last) < dedupTTL {
		viewDedupMu.Unlock()
		return
	}
	viewDedup[key] = time.Now()
	viewDedupMu.Unlock()

	go func() {
		pv := models.PageView{
			Path:      path,
			Entity:    entity,
			EntityID:  entityID,
			IP:        ip,
			UserAgent: c.Get("User-Agent"),
		}
		database.DB.Create(&pv)
	}()
}

func TrackPageView() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		path := c.Path()

		if strings.HasPrefix(path, "/admin") ||
			strings.HasPrefix(path, "/static") ||
			strings.HasPrefix(path, "/uploads") ||
			strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/ghchart") ||
			strings.HasPrefix(path, "/github-") ||
			path == "/health" ||
			path == "/login" ||
			path == "/register" ||
			path == "/logout" {
			return err
		}

		if strings.HasPrefix(path, "/posts/") && c.Method() == "GET" {
			slug := strings.TrimPrefix(path, "/posts/")
			if !strings.Contains(slug, "/") {
				trackPageView(c, "post", slug)
			}
		} else if strings.HasPrefix(path, "/projects/") && c.Method() == "GET" {
			slug := strings.TrimPrefix(path, "/projects/")
			if !strings.Contains(slug, "/") {
				trackPageView(c, "project", slug)
			}
		} else if strings.HasPrefix(path, "/service/") && c.Method() == "GET" {
			slug := strings.TrimPrefix(path, "/service/")
			if !strings.Contains(slug, "/") {
				trackPageView(c, "service", slug)
			}
		} else if path == "/" || path == "/about" || path == "/contact" || path == "/services" || path == "/projects" || path == "/posts" {
			trackPageView(c, "page", path)
		}

		return err
	}
}
