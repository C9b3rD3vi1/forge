package handlers

import (
	"math"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Frontend handlers for services
func ServiceList(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	perPage := 6
	offset := (page - 1) * perPage

	search := c.Query("search")
	category := c.Query("category")
	status := c.Query("status")

	base := database.DB.Where("published = ?", true)
	if search != "" {
		base = base.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		base = base.Where("category = ?", category)
	}
	if status != "" {
		base = base.Where("status = ?", status)
	}

	var total int64
	base.Model(&models.Services{}).Count(&total)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	var services []models.Services
	if err := base.Model(&models.Services{}).Offset(offset).Limit(perPage).Preload("TechStacks").Order("featured desc, created_at desc").Find(&services).Error; err != nil {
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Internal Server Error",
		})
	}

	return c.Render("pages/services", fiber.Map{
		"Services":     services,
		"Admin":        false,
		"CurrentPage":  page,
		"TotalPages":   totalPages,
		"TotalResults": total,
		"SearchQuery":  search,
		"Category":     category,
		"Status":       status,
		"PrevPage":     page - 1,
		"NextPage":     page + 1,
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

	database.DB.Model(&models.Services{}).Where("id = ?", service.ID).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	return c.Render("pages/service_view", fiber.Map{
		"Service": service,
		"Admin":   false,
	})
}
