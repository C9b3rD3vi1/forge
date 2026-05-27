// middleware/auth.go
package middleware

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequireAdminAuth(c *fiber.Ctx) error {
    sess, err := config.Store.Get(c)
    if err != nil {
        return c.Redirect("/admin/login")
    }

    // Must match session name
    idVal := sess.Get("user_id")
    if idVal == nil {
        return c.Redirect("/admin/login")
    }

    // Convert to UUID
    userID, err := uuid.Parse(idVal.(string))
    if err != nil {
        return c.Redirect("/admin/login")
    }

    // Fetch admin
    var user models.User
    if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
        return c.Redirect("/admin/login")
    }

    // Check admin flag
    if !user.IsAdmin {
        return c.SendStatus(fiber.StatusForbidden)
    }

    // Preserve user in context
    c.Locals("user", &user)

    return c.Next()
}



func AdminAuthMiddleware(c *fiber.Ctx) error {
    user := c.Locals("user")
    if user == nil {
        return c.Redirect("/login")
    }

    u := user.(*models.User)
    if !u.IsAdmin {
        return c.Status(403).SendString("Forbidden")
    }

    return c.Next()
}


// LogoutUser handles user logout
func LogoutUser(c *fiber.Ctx) error {
    sess, err := config.Store.Get(c)
    if err != nil {
        return err
    }
    if err := sess.Destroy(); err != nil {
        return err
    }
    return c.Redirect("/admin/login")
}
