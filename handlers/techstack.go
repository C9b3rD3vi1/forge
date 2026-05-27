package handlers

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// --- LIST ---
func AdminTechStackList(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var techStacks []models.TechStack
	if err := database.DB.Order("created_at desc").Find(&techStacks).Error; err != nil {
		return c.Status(500).SendString("Error fetching tech stacks")
	}

	return c.Render("admin/techstacks", fiber.Map{
		"Title":      "Tech Stack Management",
		"Admin":      admin,
		"TechStacks": techStacks,
	})
}

// --- CREATE FORM ---
func AdminNewTechStackForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}
	return c.Render("admin/new_techstack", fiber.Map{
		"Title": "Add Tech Stack",
		"Admin": admin,
	})
}

// --- CREATE ACTION ---
func AdminCreateTechStack(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	iconURL, err := utils.UploadImage(c, "icon")
	if err != nil {
		return c.Status(500).Render("error/500", fiber.Map{
			"Error": "Error uploading icon",
			"Admin": admin,
		})
	}

	name := c.FormValue("name")
	if name == "" {
		return c.Render("admin/new_techstack", fiber.Map{
			"Error": "Tech Stack name is required",
			"Admin": admin,
		})
	}

	tech := models.TechStack{
		Name:    name,
		IconURL: iconURL,
	}

	if err := database.DB.Create(&tech).Error; err != nil {
		return c.Status(500).Render("error/500", fiber.Map{
			"Error": "Error saving tech stack",
			"Admin": admin,
		})
	}

	return c.Redirect("/admin/techstacks")
}

// --- EDIT FORM ---
func AdminEditTechStackForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	idStr := c.Params("id")
	techID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).Render("error/400", fiber.Map{
			"Error": "Invalid Tech Stack ID",
			"Admin": admin,
		})
	}

	var tech models.TechStack
	if err := database.DB.First(&tech, "id = ?", techID).Error; err != nil {
		return c.Status(404).Render("error/404", fiber.Map{
			"Error": "Tech Stack not found",
			"Admin": admin,
		})
	}

	return c.Render("admin/edit_techstack", fiber.Map{
		"TechStack": tech,
		"Admin":     admin,
	})
}

// --- UPDATE ACTION ---
func AdminUpdateTechStack(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	idStr := c.Params("id")
	techID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).SendString("Invalid Tech Stack ID")
	}

	var tech models.TechStack
	if err := database.DB.First(&tech, "id = ?", techID).Error; err != nil {
		return c.Status(404).SendString("Tech stack not found")
	}

	iconURL, _ := utils.UploadImage(c, "icon")

	tech.Name = c.FormValue("name")
	if iconURL != "" {
		tech.IconURL = iconURL
	}

	if err := database.DB.Save(&tech).Error; err != nil {
		return c.Status(500).SendString("Error updating tech stack")
	}

	return c.Redirect("/admin/techstacks")
}

// --- DELETE ---
func AdminDeleteTechStack(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	idStr := c.Params("id")
	techID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).Render("error/400", fiber.Map{
			"Error": "Invalid Tech Stack ID",
			"Admin": admin,
		})
	}

	if err := database.DB.Delete(&models.TechStack{}, "id = ?", techID).Error; err != nil {
		return c.Status(500).Render("error/500", fiber.Map{
			"Error": "Error deleting tech stack",
			"Admin": admin,
		})
	}

	return c.Redirect("/admin/techstacks")
}
