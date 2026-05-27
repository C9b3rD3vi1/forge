package handlers

import (
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
)

// HomePageHandler renders the homepage
func HomePageHandler(c *fiber.Ctx) error {
	return c.Render("pages/index", fiber.Map{})
}

// UserRegisterHandlerForm renders the registration page
func UserRegisterHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/register", fiber.Map{})
}

// UserLoginHandlerForm renders the login page
func UserLoginHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/login", fiber.Map{})
}

// UserContactHandlerForm renders the contact page with flash messages
func UserContactHandlerForm(c *fiber.Ctx) error {
	return c.Render("pages/contact", fiber.Map{
		"Success": utils.GetFlash(c, "success"),
		"Error":   utils.GetFlash(c, "error"),
	})
}

// AboutUsHandler renders the about page
func AboutUsHandler(c *fiber.Ctx) error {
	return c.Render("pages/about", fiber.Map{})
}

// ContactForm represents the contact form structure
type ContactForm struct {
	Name    string `form:"name"`
	Email   string `form:"email"`
	Subject string `form:"subject"`
	Message string `form:"message"`
}

// UserContactHandler processes the contact form submission
func UserContactHandler(c *fiber.Ctx) error {
	// Initialize session
	sess, err := config.Store.Get(c)
	if err != nil {
		return c.Status(500).Render("error/500", fiber.Map{
			"Error":   utils.GetFlash(c, "error"),
		})
	}

	// Parse form
	form := new(ContactForm)
	if err := c.BodyParser(form); err != nil {
		sess.Set("error", "Invalid form data")
		sess.Save()
		return c.Redirect("/contact")
	}

	// Validation
	if form.Name == "" || form.Email == "" || form.Subject == "" || form.Message == "" {
		sess.Set("error", "All fields are required")
		sess.Save()
		return c.Redirect("/contact")
	}

	// Create ContactMessage model
	message := models.ContactMessage{
		Name:    form.Name,
		Email:   form.Email,
		Subject: form.Subject,
		Message: form.Message,
	}

	// Save to database
	if err := database.DB.Create(&message).Error; err != nil {
		sess.Set("error", "Failed to save message")
		sess.Save()
		return c.Redirect("/contact")
	}

	// Optionally, send email notification here (future enhancement)
	// utils.SendEmail(message.Email, message.Subject, message.Message)

	// Set success flash message
	sess.Set("success", "Your message has been sent!")
	sess.Save()

	return c.Redirect("/contact")
}
