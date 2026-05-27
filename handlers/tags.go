package handlers

import (
	"strings"
	//"log"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AdminListTags lists all tags
func AdminListTags(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var tags []models.Tag
	database.DB.Order("name asc").Find(&tags)

	return c.Render("admin/tags", fiber.Map{
		"Tags": tags,
	})
}

// AdminCreateTag creates a new tag
func AdminCreateTag(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.Status(400).SendString("Tag name is required")
	}

	tag := models.Tag{Name: name}
	if err := database.DB.Create(&tag).Error; err != nil {
		//log.Println(err)
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Error creating tag",
		})
	}

	return c.Redirect("/admin/tags")
}

// AdminDeleteTag deletes a tag by UUID
func AdminDeleteTag(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	idStr := c.Params("id")
	tagID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).SendString("Invalid tag ID")
	}

	var tag models.Tag
	if err := database.DB.First(&tag, "id = ?", tagID).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Tag not found",
		})
	}

	if err := database.DB.Delete(&tag).Error; err != nil {
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Error deleting tag",
		})
	}

	return c.Redirect("/admin/tags")
}
