package handlers

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/middleware"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

// AdminUserList shows all users
func AdminUserList(c *fiber.Ctx) error {
    var users []models.User
    database.DB.Find(&users)

    return c.Render("admin/users", fiber.Map{
        "Title": "Manage Users",
        "Users": users,
    })
}

// AdminProfile shows admin profile
func AdminProfile(c *fiber.Ctx) error {
    user := config.GetCurrentUser(c)
    return c.Render("admin/profile", fiber.Map{
        "Title": "Admin Profile",
        "User":  user,
    })
}

// AdminCreateUser handles user creation
func AdminCreateUser(c *fiber.Ctx) error {
    if c.Method() == "POST" {
        user := models.User{
            Username: c.FormValue("username"),
            Email:    c.FormValue("email"),
            Password: c.FormValue("password"),
            IsAdmin:  c.FormValue("is_admin") == "on",
            IsActive: c.FormValue("is_active") == "on",
        }

        if err := user.HashPassword(); err != nil {
            return c.Render("admin/user_form", fiber.Map{
                "Title": "Create User",
                "Error": "Error hashing password",
                "User":  user,
            })
        }

        if err := database.DB.Create(&user).Error; err != nil {
            return c.Render("admin/user_form", fiber.Map{
                "Title": "Create User",
                "Error": "Error creating user",
                "User":  user,
            })
        }

        return c.Redirect("/admin/users")
    }

    return c.Render("admin/user_form", fiber.Map{
        "Title": "Create User",
    })
}

// AdminLogout handles admin logout
func AdminLogout(c *fiber.Ctx) error {
    if err := middleware.LogoutUser(c); err != nil {
        return c.Status(500).SendString("Logout error")
    }
    return c.Redirect("/admin/login")
}
