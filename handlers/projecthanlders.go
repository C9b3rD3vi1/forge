package handlers

import (
	"log"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

// ProjectList displays all projects (public)
func ProjectList(c *fiber.Ctx) error {
	var projects []models.Projects

	// Fetch projects with related tech stacks
	if err := database.DB.Preload("TechStacks").Order("created_at desc").Find(&projects).Error; err != nil {
		log.Printf("Error fetching projects: %v", err)
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Internal Server Error",
		})
	}

	return c.Render("pages/projects", fiber.Map{
		"Title":    "Projects",
		"Projects": projects,
	})
}

// ProjectView displays a single project by slug (public)
func ProjectView(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var project models.Projects

	// Fetch project with tech stacks by slug
	if err := database.DB.Preload("TechStacks").Where("slug = ?", slug).First(&project).Error; err != nil {
		log.Printf("Project not found for slug '%s': %v", slug, err)
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Project not found",
		})
	}

	return c.Render("pages/project_view", fiber.Map{
		"Title":   project.Title,
		"Project": project,
	})
}
