package handlers

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func AdminSettings(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var settings []models.Setting
	database.DB.Order("key asc").Find(&settings)

	settingMap := make(map[string]string)
	for _, s := range settings {
		settingMap[s.Key] = s.Value
	}

	return c.Render("admin/settings", fiber.Map{
		"Title":    "Site Settings",
		"Admin":    admin,
		"Settings": settingMap,
		"Success":  utils.GetFlash(c, "success"),
		"Error":    utils.GetFlash(c, "error"),
	})
}

func AdminSettingsUpdate(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	keys := []string{
		"site_name", "site_tagline", "site_description",
		"github_username", "twitter_url", "linkedin_url",
		"hero_title", "hero_subtitle",
		"contact_email", "site_keywords",
		"open_source_projects",
		"stat_faster_deploys", "stat_uptime", "stat_certifications", "stat_yrs_production",
	}

	for _, key := range keys {
		value := c.FormValue(key)
		var setting models.Setting
		result := database.DB.Where("key = ?", key).First(&setting)
		if result.Error != nil {
			database.DB.Create(&models.Setting{Key: key, Value: value})
		} else {
			database.DB.Model(&setting).Update("value", value)
		}
	}

	utils.SetFlash(c, "success", "Settings updated successfully")
	return c.Redirect("/admin/settings")
}

func AdminProfileUpdate(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	admin.Username = c.FormValue("username")
	admin.Email = c.FormValue("email")
	admin.FullName = c.FormValue("full_name")

	if err := database.DB.Save(admin).Error; err != nil {
		utils.SetFlash(c, "error", "Failed to update profile")
		return c.Redirect("/admin/settings")
	}

	utils.SetFlash(c, "success", "Profile updated successfully")
	return c.Redirect("/admin/settings")
}

func AdminPasswordUpdate(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	current := c.FormValue("current_password")
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(current)); err != nil {
		utils.SetFlash(c, "error", "Current password is incorrect")
		return c.Redirect("/admin/settings")
	}

	newPass := c.FormValue("new_password")
	confirm := c.FormValue("confirm_password")
	if newPass == "" {
		utils.SetFlash(c, "error", "New password cannot be empty")
		return c.Redirect("/admin/settings")
	}
	if newPass != confirm {
		utils.SetFlash(c, "error", "Passwords do not match")
		return c.Redirect("/admin/settings")
	}

	admin.Password = newPass
	if err := admin.HashPassword(); err != nil {
		utils.SetFlash(c, "error", "Failed to hash password")
		return c.Redirect("/admin/settings")
	}

	if err := database.DB.Save(admin).Error; err != nil {
		utils.SetFlash(c, "error", "Failed to update password")
		return c.Redirect("/admin/settings")
	}

	utils.SetFlash(c, "success", "Password updated successfully")
	return c.Redirect("/admin/settings")
}
