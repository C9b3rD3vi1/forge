package handlers

import (
	"strings"
	"time"

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

	stackIDs := c.FormValue("techstacks")

	var techStacks []models.TechStack
	if stackIDs != "" {
		ids := strings.Split(stackIDs, ",")
		database.DB.Where("id IN ?", ids).Find(&techStacks)
	}

	imageURL, _ := utils.UploadImage(c, "image")
	slug := utils.UniqueSlug(database.DB, "services", c.FormValue("title"))

	// Parse published_at
	var publishedAt time.Time
	if pa := c.FormValue("published_at"); pa != "" {
		publishedAt, _ = time.Parse("2006-01-02", pa)
	}

	service := models.Services{
		Title:            c.FormValue("title"),
		Description:      c.FormValue("description"),
		Slug:             slug,
		ImageURL:         imageURL,
		Category:         c.FormValue("category"),
		Tags:             c.FormValue("tags"),
		Featured:         c.FormValue("featured") == "on",
		Published:        c.FormValue("published") == "on",
		Status:           c.FormValue("status"),
		PublishedAt:      publishedAt,
		MetaDescription:  c.FormValue("meta_description"),
		CanonicalURL:     c.FormValue("canonical_url"),
		TechStacks:       techStacks,
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

	serviceTechStackIDs := make(map[string]bool)
	for _, ts := range service.TechStacks {
		serviceTechStackIDs[ts.ID.String()] = true
	}

	return c.Render("admin/edit_service", fiber.Map{
		"Title":               "Edit Service",
		"Admin":               admin,
		"Service":             service,
		"TechStacks":          allTechStacks,
		"ServiceTechStackIDs": serviceTechStackIDs,
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

    id := c.Params("id")
    var service models.Services
    if err := database.DB.Where("id = ?", id).First(&service).Error; err != nil {
        return c.Status(404).SendString("Service not found")
    }

    service.Title = c.FormValue("title")
    service.Description = c.FormValue("description")
    service.Category = c.FormValue("category")
    service.Tags = c.FormValue("tags")
    service.Featured = c.FormValue("featured") == "on"
    service.Published = c.FormValue("published") == "on"
    service.Status = c.FormValue("status")
    service.MetaDescription = c.FormValue("meta_description")
    service.CanonicalURL = c.FormValue("canonical_url")

    if pa := c.FormValue("published_at"); pa != "" {
        if parsed, err := time.Parse("2006-01-02", pa); err == nil {
            service.PublishedAt = parsed
        }
    }

    if imageURL, _ := utils.UploadImage(c, "image"); imageURL != "" {
        service.ImageURL = imageURL
    }

    service.Slug = utils.UniqueSlug(database.DB, "services", service.Title)

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