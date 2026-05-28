package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/znbang/gofiber-layout/html"
)

// LayoutMiddleware dynamically sets the layout template based on the route path.
// If the path starts with "/admin", it uses the admin layout; otherwise, the public layout.
// DynamicLayout automatically switches layout templates based on URL prefix
// DynamicLayoutMiddleware sets the layout template based on the route path.
func DynamicLayoutMiddleware(engine *html.Engine) fiber.Handler {
	return func(c *fiber.Ctx) error {
		path := c.Path()

		// ✅ Skip layout for admin login and 2FA verify pages
		if path == "/admin/login" || path == "/admin/register" || path == "/admin/verify-otp" {
			engine.Layout("") // no layout
			return c.Next()
		}

		// ✅ Apply layout depending on route
		if strings.HasPrefix(path, "/admin") {
			engine.Layout("layouts/admin")
		} else {
			engine.Layout("layouts/base")
		}

		return c.Next()
	}
}