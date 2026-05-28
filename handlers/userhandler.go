package handlers

import (
	"html/template"
	"os"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
)

func HomePageHandler(c *fiber.Ctx) error {
	return c.Render("pages/index", fiber.Map{})
}

func UserRegisterHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/register", fiber.Map{})
}

func UserLoginHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{})
}

func UserContactHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/contact", fiber.Map{
		"Success": utils.GetFlash(c, "success"),
		"Error":   utils.GetFlash(c, "error"),
	})
}

func AboutUsHandler(c *fiber.Ctx) error {
	return c.Render("pages/about", fiber.Map{})
}

type ContactForm struct {
	Name    string `form:"name"`
	Email   string `form:"email"`
	Subject string `form:"subject"`
	Message string `form:"message"`
}

func UserContactHandler(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return c.Status(500).Render("error/500", fiber.Map{
			"Error": utils.GetFlash(c, "error"),
		})
	}

	form := new(ContactForm)
	if err := c.BodyParser(form); err != nil {
		sess.Set("error", "Invalid form data")
		sess.Save()
		return c.Redirect("/contact")
	}

	if form.Name == "" || form.Email == "" || form.Subject == "" || form.Message == "" {
		sess.Set("error", "All fields are required")
		sess.Save()
		return c.Redirect("/contact")
	}

	message := models.ContactMessage{
		Name:    form.Name,
		Email:   form.Email,
		Subject: form.Subject,
		Message: form.Message,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		sess.Set("error", "Failed to save message")
		sess.Save()
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
			MessageBody:    escapedMsg,
			LinkURL:        siteURL + "/admin/contacts/" + message.ID.String(),
		})
	}

	sess.Set("success", "Your message has been sent!")
	sess.Save()

	return c.Redirect("/contact")
}
