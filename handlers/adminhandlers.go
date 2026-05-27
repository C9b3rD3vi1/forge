package handlers

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

// AdminDashboard renders the admin dashboard
func AdminDashboard(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil {
        return c.Redirect("/admin/login")
    }

    if !admin.IsAdmin {
        return c.SendStatus(fiber.StatusForbidden)
    }

    var posts []models.Post
    var projects []models.Projects
    var services []models.Services
    var users []models.User
    var messages []models.ContactMessage

    database.DB.Order("created_at desc").Find(&posts)
    database.DB.Order("created_at desc").Find(&projects)
    database.DB.Order("created_at desc").Find(&users)
    database.DB.Order("created_at desc").Find(&services)
    database.DB.Order("created_at desc").Find(&messages)

    return c.Render("admin/dashboard", fiber.Map{
        "Title":    "Admin Dashboard",
        "Admin":    admin, // whole struct → you can display Username, Email, etc.
        "Posts":    posts,
        "Projects": projects,
        "Services": services,
        "Users":    users,
        "Messages": messages,
    })
}

