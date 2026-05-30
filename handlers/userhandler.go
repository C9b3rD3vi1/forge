package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
)

type OSSProject struct {
	Icon string `json:"icon"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

func HomePageHandler(c *fiber.Ctx) error {
	var projects []models.Projects
	database.DB.Preload("TechStacks").Where("published = ? AND featured = ?", true, true).Order("created_at desc").Find(&projects)

	var githubUser string
	database.DB.Model(&models.Setting{}).Where("key = ?", "github_username").Select("value").Scan(&githubUser)
	if githubUser == "" {
		githubUser = "C9b3rD3vi1"
	}

	var ossRaw string
	database.DB.Model(&models.Setting{}).Where("key = ?", "open_source_projects").Select("value").Scan(&ossRaw)
	var ossProjects []OSSProject
	if ossRaw != "" {
		json.Unmarshal([]byte(ossRaw), &ossProjects)
	}
	if ossProjects == nil {
		ossProjects = []OSSProject{}
	}

	ghStats := utils.FetchGitHubStats(githubUser)

	var projectCount int64
	database.DB.Model(&models.Projects{}).Where("published = ?", true).Count(&projectCount)
	liveSystems := fmt.Sprintf("%d+", projectCount)

	var fasterDeploys, yrsProd, certs string
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_faster_deploys").Select("value").Scan(&fasterDeploys)
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_yrs_production").Select("value").Scan(&yrsProd)
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_certifications").Select("value").Scan(&certs)

	if fasterDeploys == "" {
		fasterDeploys = "5×"
	}
	if yrsProd == "" {
		yrsProd = "3+"
	}
	if certs == "" {
		certs = "7"
	}

	return c.Render("pages/index", fiber.Map{
		"Projects":         projects,
		"GitHubUsername":   githubUser,
		"GitHubProfileURL": "https://github.com/" + githubUser,
		"GitHubReposURL":   "https://github.com/" + githubUser + "?tab=repositories",
		"OSSProjects":      ossProjects,
		"GHStats":          ghStats,
		"LiveSystems":      liveSystems,
		"FasterDeploys":    fasterDeploys,
		"YrsProduction":    yrsProd,
		"Certifications":   certs,
	})
}

func UserRegisterHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/register", fiber.Map{})
}

func UserLoginHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{
		"Redirect": c.Query("redirect"),
	})
}

func UserContactHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/contact", fiber.Map{
		"Success": utils.GetFlash(c, "success"),
		"Error":   utils.GetFlash(c, "error"),
	})
}

func AboutUsHandler(c *fiber.Ctx) error {
	var projectCount int64
	database.DB.Model(&models.Projects{}).Where("published = ?", true).Count(&projectCount)

	var fasterDeploys, uptime, certs string
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_faster_deploys").Select("value").Scan(&fasterDeploys)
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_uptime").Select("value").Scan(&uptime)
	database.DB.Model(&models.Setting{}).Where("key = ?", "stat_certifications").Select("value").Scan(&certs)

	if fasterDeploys == "" {
		fasterDeploys = "5×"
	}
	if uptime == "" {
		uptime = "99.9%"
	}
	if certs == "" {
		certs = "7"
	}

	liveSystems := fmt.Sprintf("%d+", projectCount)

	return c.Render("pages/about", fiber.Map{
		"LiveSystems":    liveSystems,
		"FasterDeploys":  fasterDeploys,
		"Uptime":         uptime,
		"Certifications": certs,
	})
}

type ContactForm struct {
	Name     string   `form:"name"`
	Email    string   `form:"email"`
	Subject  string   `form:"subject"`
	Message  string   `form:"message"`
	Services []string `form:"services"`
}

func UserContactHandler(c *fiber.Ctx) error {
	form := new(ContactForm)
	if err := c.BodyParser(form); err != nil {
		utils.SetFlash(c, "error", "Invalid form data")
		return c.Redirect("/contact")
	}

	if form.Name == "" || form.Email == "" || form.Subject == "" || form.Message == "" {
		utils.SetFlash(c, "error", "All fields are required")
		return c.Redirect("/contact")
	}

	servicesStr := ""
	if len(form.Services) > 0 {
		servicesStr = strings.Join(form.Services, ", ")
	}

	message := models.ContactMessage{
		Name:     form.Name,
		Email:    form.Email,
		Subject:  form.Subject,
		Message:  form.Message,
		Services: servicesStr,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		utils.SetFlash(c, "error", "Failed to save message")
		return c.Redirect("/contact")
	}

	siteURL := utils.GetEnv("SITE_URL", "http://localhost:3031")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	escapedMsg := template.HTML(utils.EscapeHTML(form.Message))

	utils.SendEmailAsync(form.Email, "Thank you for contacting Forge.Hub", "auto_reply.html", utils.EmailData{
		RecipientName:  form.Name,
		RecipientEmail: form.Email,
		MessageBody:    escapedMsg,
	})

	if adminEmail != "" {
		utils.SendEmailAsync(adminEmail, "New Contact Message: "+form.Subject, "admin_notify.html", utils.EmailData{
			RecipientName:  form.Name,
			RecipientEmail: form.Email,
			Subject:        form.Subject,
			Services:       servicesStr,
			MessageBody:    escapedMsg,
			LinkURL:        siteURL + "/admin/contacts/" + message.ID.String(),
		})
	}

	utils.SetFlash(c, "success", "Your message has been sent!")

	return c.Redirect("/contact")
}
