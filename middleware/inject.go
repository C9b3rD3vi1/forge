package middleware

import (
	"strings"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

func InjectGlobalData() fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Skip for non-HTML routes
		if strings.HasPrefix(path, "/static") || strings.HasPrefix(path, "/uploads") ||
			strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/ghchart") ||
			strings.HasPrefix(path, "/github-") || strings.HasPrefix(path, "/admin") {
			return c.Next()
		}

		var services []models.Services
		database.DB.Where("published = ?", true).Order("created_at desc").Find(&services)
		if services == nil {
			services = []models.Services{}
		}

		c.Locals("FooterServices", services)

		// Check if user is logged in
		sess, err := config.Store.Get(c)
		if err == nil {
			userID := sess.Get("user_id")
			c.Locals("IsLoggedIn", userID != nil)
		} else {
			c.Locals("IsLoggedIn", false)
		}

		return c.Next()
	}
}
