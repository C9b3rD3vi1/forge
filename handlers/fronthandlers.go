package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/C9b3rD3vi1/forge/models"
    
    "github.com/C9b3rD3vi1/forge/database"
)

// Frontend handlers for services
func ServiceList(c *fiber.Ctx) error {
    var services []models.Services
    database.DB.Order("created_at desc").Find(&services)

    return c.Render("pages/services", fiber.Map{
        "Services": services,
        "Admin":    false, // public page, no admin controls
    })
}

// ServiceView displays a single service view
func ServiceView(c *fiber.Ctx) error {
    slug := c.Params("slug")
    var service models.Services
    if err := database.DB.Preload("TechStacks").Where("slug = ?", slug).First(&service).Error; err != nil {
        return c.Status(404).Render("errors/404", fiber.Map{
            "Message": "Service not found",
        })
    }

    return c.Render("pages/service_view", fiber.Map{
        "Service": service,
        "Admin":   false,
    })
}
