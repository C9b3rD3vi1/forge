package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRegisterHandler handles new user registration
func UserRegisterHandler(c *fiber.Ctx) error {
	fullname := strings.TrimSpace(c.FormValue("fullname"))
	username := strings.TrimSpace(c.FormValue("username"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))
	passwordConfirm := strings.TrimSpace(c.FormValue("password_confirm"))

	// Validate
	if fullname == "" || username == "" || email == "" || password == "" || passwordConfirm == "" {
		return c.Render("pages/register", fiber.Map{"Error": "All fields are required"})
	}
	if password != passwordConfirm {
		return c.Render("pages/register", fiber.Map{"Error": "Passwords do not match"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).SendString("Error hashing password")
	}

	// Check if email or username already exists
	var existing models.User
	if err := database.DB.Where("email = ? OR username = ?", email, username).First(&existing).Error; err == nil {
		return c.Render("pages/register", fiber.Map{"Error": "Email or username already taken"})
	}

	user := models.User{
		FullName: fullname,
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).SendString("Error creating user")
	}

	return c.Redirect("/login")
}

// UserLoginHandler handles user login
func UserLoginHandler(c *fiber.Ctx) error {
	email := strings.TrimSpace(c.FormValue("email"))
	password := strings.TrimSpace(c.FormValue("password"))

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Render("pages/login", fiber.Map{"Error": "Invalid email or password"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return c.Render("pages/login", fiber.Map{"Error": "Invalid email or password"})
	}

	// Store session (UUID-safe)
	sess, err := config.Store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Session error")
	}

	sess.Set("user_id", user.ID.String()) // store UUID as string
	sess.Set("username", user.Username)
	sess.Set("_ip", c.IP())

	if err := sess.Save(); err != nil {
		return c.Status(500).SendString("Session save error")
	}

	redirect := c.FormValue("redirect")
	if redirect == "" {
		redirect = c.Query("redirect")
	}
	if redirect == "" {
		redirect = "/"
	}

	return c.Redirect(redirect)
}

// UserLogoutHandler logs out the user
func UserLogoutHandler(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Session error")
	}

	sess.Destroy()
	return c.Redirect("/login")
}
