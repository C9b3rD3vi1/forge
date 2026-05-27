package handlers

import (
	"strings"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
)

// --- Services ---
//// Use Preload("TechStacks") to eager load the many-to-many relationship
// AdminServiceList shows all services
func AdminServiceList(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var services []models.Services
	if err := database.DB.Preload("TechStacks").Order("created_at desc").Find(&services).Error; err != nil {
		return c.Status(500).SendString("Error loading services")
	}

	return c.Render("admin/services", fiber.Map{
		"Title":    "Manage Services",
		"Admin":    admin,
		"Services": services,
	})
}


func AdminNewServicePage(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    var techStacks []models.TechStack
    if err := database.DB.Order("created_at desc").Find(&techStacks).Error; err != nil {
        return c.Status(500).SendString("Error fetching tech stacks")
    }

    return c.Render("admin/new_service", fiber.Map{
        "Admin":      admin,
        "TechStacks": techStacks, // must match {{ .TechStacks }} in template
    })
}

// AdminCreateServices handles the creation of a new service
func AdminCreateServices(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	title := c.FormValue("title")
	description := c.FormValue("description")
	stackIDs := c.FormValue("techstacks") // comma-separated IDs

	var techStacks []models.TechStack
	if stackIDs != "" {
		ids := strings.Split(stackIDs, ",")
		database.DB.Where("id IN ?", ids).Find(&techStacks)
	}

	imageURL, _ := utils.UploadImage(c, "image")
	slug := utils.UniqueSlug(database.DB, "services", title)

	service := models.Services{
		Title:       title,
		Description: description,
		Slug:        slug,
		ImageURL:    imageURL,
		TechStacks:  techStacks,
	}

	if err := database.DB.Create(&service).Error; err != nil {
		return c.Status(500).SendString("Error saving service")
	}

	return c.Redirect("/admin/services")
}



// AdminEditServiceForm handles the form for editing a service
func AdminEditServiceForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var service models.Services
	if err := database.DB.Preload("TechStacks").Where("id = ?", id).First(&service).Error; err != nil {
		return c.Status(404).SendString("Service not found")
	}

	var allTechStacks []models.TechStack
	database.DB.Order("created_at desc").Find(&allTechStacks)

	return c.Render("admin/edit_service", fiber.Map{
		"Title":      "Edit Service",
		"Admin":      admin,
		"Service":    service,
		"TechStacks": allTechStacks,
	})
}

// AdminDeleteService handles the deletion of a service

// AdminDeleteService handles deletion
func AdminDeleteService(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	if err := database.DB.Delete(&models.Services{}, "id = ?", id).Error; err != nil {
		return c.Status(500).SendString("Error deleting service")
	}

	return c.Redirect("/admin/services")
}

// AdminViewService shows a single service
func AdminViewService(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	slug := c.Params("slug")
	var service models.Services
	if err := database.DB.Preload("TechStacks").Where("slug = ?", slug).First(&service).Error; err != nil {
		return c.Status(404).SendString("Service not found")
	}

	return c.Render("admin/view_service", fiber.Map{
		"Title":   service.Title,
		"Admin":   admin,
		"Service": service,
	})
}

// AdminUpdateService handles the update of a service
func AdminUpdateService(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    slug := c.Params("slug")
    var service models.Services
    if err := database.DB.Where("slug = ?", slug).First(&service).Error; err != nil {
        return c.Status(404).SendString("Service not found")
    }

    // Update fields
    service.Title = c.FormValue("title")
    service.Description = c.FormValue("description")

    if imageURL, _ := utils.UploadImage(c, "image"); imageURL != "" {
        service.ImageURL = imageURL
    }

    // Update slug
    service.Slug = utils.UniqueSlug(database.DB, "services", service.Title)

    // Update TechStacks
    stackIDs := c.FormValue("techstacks")
    if stackIDs != "" {
        ids := strings.Split(stackIDs, ",")
        var techStacks []models.TechStack
        database.DB.Where("id IN ?", ids).Find(&techStacks)
        database.DB.Model(&service).Association("TechStacks").Replace(techStacks)
    } else {
        database.DB.Model(&service).Association("TechStacks").Clear()
    }

    if err := database.DB.Save(&service).Error; err != nil {
        return c.Status(500).SendString("Error updating service")
    }

    return c.Redirect("/admin/services")
}



// AdminNewServiceForm renders the form for creating a new service
func AdminNewServiceForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var techStacks []models.TechStack
	database.DB.Order("created_at desc").Find(&techStacks)

	return c.Render("admin/new_service", fiber.Map{
		"Title":      "Add New Service",
		"Admin":      admin,
		"TechStacks": techStacks,
	})
}