package auth

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/crypto/bcrypt"
)

// AdminLoginForm renders the admin login page
func AdminLoginForm(c *fiber.Ctx) error {
	fmt.Println(" Rendering admin login page")
	return c.Render("admin/login", fiber.Map{})
}

// AdminAuthHandler handles admin login
func AdminAuthHandler(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	remember := c.FormValue("remember") == "on"

	fmt.Printf(" Admin login attempt: email=%s, remember=%v\n", email, remember)

	var admin models.User
	if err := database.DB.Where("email = ? AND is_admin = ?", email, true).First(&admin).Error; err != nil {
		fmt.Printf(" Admin not found for email: %s\n", email)
		return c.Status(401).Render("admin/login", fiber.Map{"Error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		fmt.Printf(" Password mismatch for admin email: %s\n", email)
		return c.Status(401).Render("admin/login", fiber.Map{"Error": "Invalid credentials"})
	}

	fmt.Printf(" Password correct for admin: %s (ID=%s)\n", email, admin.ID.String())

	sess, err := config.Store.Get(c)
	if err != nil {
		fmt.Println(" Session error:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	if admin.TwoFASecret != "" {
		fmt.Printf(" 2FA required for admin: %s\n", email)
		sess.Set("2fa_user_id", admin.ID.String())
		sess.SetExpiry(10 * time.Minute)
		if err := sess.Save(); err != nil {
			fmt.Println(" Error saving session for 2FA:", err)
			return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
		}
		return c.Redirect("/admin/verify-otp")
	}

	fmt.Printf(" Admin login successful: %s\n", email)
	sess.Set("user_id", admin.ID.String())
	if remember {
		sess.SetExpiry(48 * time.Hour)
		fmt.Println(" Remember me enabled, session expiry extended")
	}
	if err := sess.Save(); err != nil {
		fmt.Println(" Error saving session:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	c.Locals("user", &admin)
	return c.Redirect("/admin/dashboard")
}

// ShowOTPPage renders the OTP verification page
func ShowOTPPage(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil || sess.Get("2fa_user_id") == nil {
		return c.Redirect("/admin/login")
	}
	return c.Render("admin/verify", fiber.Map{})
}

// VerifyOTPHandler verifies 2FA OTP with brute-force protection
func VerifyOTPHandler(c *fiber.Ctx) error {
	sess, err := config.Store.Get(c)
	if err != nil {
		fmt.Println(" Session error while verifying OTP:", err)
		return c.Redirect("/admin/login")
	}

	userID := sess.Get("2fa_user_id")
	if userID == nil {
		fmt.Println(" No 2FA user ID in session, redirecting to login")
		return c.Redirect("/admin/login")
	}

	// Brute-force protection: track attempts
	attempts := 0
	if a := sess.Get("2fa_attempts"); a != nil {
		attempts = a.(int)
	}
	if attempts >= 5 {
		fmt.Printf(" 2FA brute-force lockout for user ID=%v\n", userID)
		return c.Render("admin/verify", fiber.Map{"Error": "Too many attempts. Try again in 5 minutes."})
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		fmt.Printf(" User not found in DB for 2FA ID: %v\n", userID)
		return c.Redirect("/admin/login")
	}

	otpCode := c.FormValue("otp")
	if !totp.Validate(otpCode, user.TwoFASecret) {
		attempts++
		sess.Set("2fa_attempts", attempts)
		if attempts >= 5 {
			sess.SetExpiry(5 * time.Minute)
		}
		sess.Save()
		fmt.Printf(" Invalid OTP for user ID=%s (attempt %d/5)\n", user.ID.String(), attempts)
		return c.Render("admin/verify", fiber.Map{"Error": "Invalid OTP"})
	}

	// Reset attempts on success
	sess.Delete("2fa_attempts")

	fmt.Printf(" OTP verified for admin ID=%s\n", user.ID.String())
	sess.Delete("2fa_user_id")
	sess.Set("user_id", user.ID.String())
	sess.SetExpiry(24 * time.Hour)
	if err := sess.Save(); err != nil {
		fmt.Println(" Error saving session after OTP verification:", err)
		return c.Status(500).Render("error/500", fiber.Map{"Error": "Session error"})
	}

	c.Locals("user", &user)
	return c.Redirect("/admin/dashboard")
}

// AdminSetup2FA generates TOTP secret and shows QR code
func AdminSetup2FA(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	if admin.TwoFASecret != "" {
		return c.Redirect("/admin/settings")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Forge.Hub",
		AccountName: admin.Email,
	})
	if err != nil {
		return c.Status(500).SendString("Failed to generate TOTP key")
	}

	// Save temporary secret in session
	sess, _ := config.Store.Get(c)
	sess.Set("2fa_temp_secret", key.Secret())
	sess.Save()

	// Generate QR code PNG
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return c.Status(500).SendString("Failed to generate QR code")
	}

	qrDataURI := template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(png))

	return c.Render("admin/setup_2fa", fiber.Map{
		"Title":    "Setup Two-Factor Authentication",
		"Admin":    admin,
		"Secret":   key.Secret(),
		"QRData":   qrDataURI,
		"Error":    c.Query("error"),
	})
}

// AdminConfirm2FA validates OTP and saves the TOTP secret
func AdminConfirm2FA(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	sess, _ := config.Store.Get(c)
	tempSecret, ok := sess.Get("2fa_temp_secret").(string)
	if !ok || tempSecret == "" {
		return c.Redirect("/admin/settings/2fa/setup")
	}

	otpCode := c.FormValue("otp")
	if !totp.Validate(otpCode, tempSecret) {
		utils.SetFlash(c, "error", "Invalid OTP. Please try again.")
		return c.Redirect("/admin/settings/2fa/setup")
	}

	admin.TwoFASecret = tempSecret
	if err := database.DB.Save(admin).Error; err != nil {
		fmt.Println(" Error saving TOTP secret:", err)
		return c.Status(500).SendString("Failed to save TOTP secret")
	}

	sess.Delete("2fa_temp_secret")
	sess.Save()

	utils.SetFlash(c, "success", "Two-factor authentication has been enabled.")
	return c.Redirect("/admin/settings")
}

// AdminDisable2FA clears the TOTP secret after password confirmation
func AdminDisable2FA(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	password := c.FormValue("password")
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		utils.SetFlash(c, "error", "Incorrect password. 2FA was not disabled.")
		return c.Redirect("/admin/settings")
	}

	admin.TwoFASecret = ""
	if err := database.DB.Save(admin).Error; err != nil {
		return c.Status(500).SendString("Failed to disable 2FA")
	}

	utils.SetFlash(c, "success", "Two-factor authentication has been disabled.")
	return c.Redirect("/admin/settings")
}

// AdminLogoutHandler logs out the admin
func AdminLogoutHandler(c *fiber.Ctx) error {
	sess, _ := config.Store.Get(c)
	fmt.Println(" Admin logout, destroying session")
	sess.Destroy()
	return c.Redirect("/admin/login")
}
