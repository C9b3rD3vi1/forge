package handlers

import (
	"log"
	"math"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ProjectList displays all projects (public)
func ProjectList(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	perPage := 6
	offset := (page - 1) * perPage

	search := c.Query("search")
	category := c.Query("category")
	status := c.Query("status")
	difficulty := c.Query("difficulty")

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
	if difficulty != "" {
		base = base.Where("difficulty = ?", difficulty)
	}

	var total int64
	base.Model(&models.Projects{}).Count(&total)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	var projects []models.Projects
	if err := base.Model(&models.Projects{}).Offset(offset).Limit(perPage).Preload("TechStacks").Order("featured desc, created_at desc").Find(&projects).Error; err != nil {
		log.Printf("Error fetching projects: %v", err)
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Internal Server Error",
		})
	}

	return c.Render("pages/projects", fiber.Map{
		"Title":        "Projects",
		"Projects":     projects,
		"CurrentPage":  page,
		"TotalPages":   totalPages,
		"TotalResults": total,
		"SearchQuery":  search,
		"Category":     category,
		"Status":       status,
		"Difficulty":   difficulty,
		"PrevPage":     page - 1,
		"NextPage":     page + 1,
	})
}

// ProjectView displays a single project by slug (public)
func ProjectView(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var project models.Projects

	if err := database.DB.Preload("TechStacks").Where("slug = ?", slug).First(&project).Error; err != nil {
		log.Printf("Project not found for slug '%s': %v", slug, err)
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Project not found",
		})
	}

	database.DB.Model(&models.Projects{}).Where("id = ?", project.ID).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	return c.Render("pages/project_view", fiber.Map{
		"Title":   project.Title,
		"Project": project,
	})
}
