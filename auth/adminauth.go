package auth

import (
	"fmt"
	"time"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

// AdminLoginForm renders the admin login page
func AdminLoginForm(c *fiber.Ctx) error {
	fmt.Println("📝 Rendering admin login page")
	return c.Render("admin/login", fiber.Map{})
}

// AdminAuthHandler handles admin login
func AdminAuthHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	remember := c.FormValue("remember") == "on"

	fmt.Printf("🔑 Admin login attempt: email=%s, remember=%v\n", email, remember)

	// Find admin user
	var admin models.User
	if err := database.DB.Where("email = ? AND is_admin = ?", email, true).First(&admin).Error; err != nil {
		fmt.Printf("❌ Admin not found for email: %s\n", email)
		return c.Status(401).Render("admin/login", fiber.Map{"Error": "Invalid credentials"})
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		fmt.Printf("❌ Password mismatch for admin email: %s\n", email)
		return c.Status(401).Render("admin/login", fiber.Map{"Error": "Invalid credentials"})
	}

	fmt.Printf("✅ Password correct for admin: %s (ID=%s)\n", email, admin.ID.String())

	// Save user ID in session
	sess, err := config.Store.Get(c)
	if err != nil {
		fmt.Println("❌ Session error:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	if admin.TwoFASecret != "" {
		// Require 2FA
		fmt.Printf("🔒 2FA required for admin: %s\n", email)
		sess.Set("2fa_user_id", admin.ID.String())
		if err := sess.Save(); err != nil {
			fmt.Println("❌ Error saving session for 2FA:", err)
			return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
		}
		return c.Redirect("/admin/verify-otp")
	}

	// Normal login
	fmt.Printf("🔓 Admin login successful: %s\n", email)
	sess.Set("user_id", admin.ID.String()) 
	if remember {
		sess.SetExpiry(48 * time.Hour)
		fmt.Println("⏳ Remember me enabled, session expiry extended")
	}
	if err := sess.Save(); err != nil {
		fmt.Println("❌ Error saving session:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	c.Locals("user", &admin)
	return c.Redirect("/admin/dashboard")
}

// ShowOTPPage renders the OTP verification page
func ShowOTPPage(c *fiber.Ctx) error {
	fmt.Println("📝 Rendering OTP verification page")
	return c.Render("admin/verify", fiber.Map{})
}

// VerifyOTPHandler verifies 2FA OTP
func VerifyOTPHandler(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		fmt.Println("❌ Session error while verifying OTP:", err)
		return c.Redirect("/admin/login")
	}

	userID := sess.Get("2fa_user_id")
	if userID == nil {
		fmt.Println("⚠️ No 2FA user ID in session, redirecting to login")
		return c.Redirect("/admin/login")
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		fmt.Printf("❌ User not found in DB for 2FA ID: %v\n", userID)
		return c.Redirect("/admin/login")
	}

	otpCode := c.FormValue("otp")
	if !totp.Validate(otpCode, user.TwoFASecret) {
		fmt.Printf("❌ Invalid OTP for user ID=%s\n", user.ID.String())
		return c.Render("admin/verify", fiber.Map{"Error": "Invalid OTP"})
	}

	// OTP verified: set full admin session
	fmt.Printf("✅ OTP verified for admin ID=%s\n", user.ID.String())
	sess.Delete("2fa_user_id")
	sess.Set("admin", user.ID.String())
	if err := sess.Save(); err != nil {
		fmt.Println("❌ Error saving session after OTP verification:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	c.Locals("user", &user)
	return c.Redirect("/admin/dashboard")
}

// AdminLogoutHandler logs out the admin
func AdminLogoutHandler(c *fiber.Ctx) error {
	sess, _ := config.Store.Get(c)
	fmt.Println("🔒 Admin logout, destroying session")
	sess.Destroy()
	return c.Redirect("/admin/login")
}
