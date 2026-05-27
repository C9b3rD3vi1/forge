package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"

    "github.com/C9b3rD3vi1/forge/database"
    "github.com/C9b3rD3vi1/forge/models"
    "github.com/C9b3rD3vi1/forge/utils"
)


// ------------------------------
// LOAD USER (UUID SAFE)
// ------------------------------
func loadUserByID(idStr string) (*models.User, error) {
    id, err := uuid.Parse(idStr)
    if err != nil {
        return nil, err
    }

    var user models.User
    if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
        return nil, err
    }

    return &user, nil
}



// ------------------------------
// EDIT USER PAGE
// ------------------------------
func AdminUserEditPage(c *fiber.Ctx) error {
    id := c.Params("id")

    user, err := loadUserByID(id)
    if err != nil {
        return c.Status(404).SendString("User not found")
    }

    return c.Render("admin/users/edit", fiber.Map{
        "Title": "Edit User",
        "User":  user,
    })
}



// ------------------------------
// UPDATE USER (SAVE EDIT)
// ------------------------------
func AdminUserEdit(c *fiber.Ctx) error {
    id := c.Params("id")

    user, err := loadUserByID(id)
    if err != nil {
        return c.Status(404).SendString("User not found")
    }

    form := new(models.User)
    if err := c.BodyParser(form); err != nil {
        return c.Status(fiber.StatusBadRequest).SendString("Invalid form data")
    }

    // Update fields
    user.FullName = form.FullName
    user.Email = form.Email
    user.Address = form.Address
    user.IsAdmin = form.IsAdmin

    // Update password only if provided
    if form.Password != "" {
        hashed, err := utils.HashPassword(form.Password)
        if err != nil {
            return c.Status(500).SendString("Error hashing password")
        }
        user.Password = hashed
    }

    database.DB.Save(user)
    return c.Redirect("/admin/users")
}



// ------------------------------
// VIEW USER
// ------------------------------
func AdminViewUser(c *fiber.Ctx) error {
    id := c.Params("id")

    user, err := loadUserByID(id)
    if err != nil {
        return c.Status(404).SendString("User not found")
    }

    return c.Render("admin/users/view", fiber.Map{
        "Title": "View User",
        "User":  user,
    })
}



// ------------------------------
// DELETE USER
// ------------------------------
func AdminDeleteUser(c *fiber.Ctx) error {
    id := c.Params("id")

    user, err := loadUserByID(id)
    if err != nil {
        return c.Status(404).SendString("User not found")
    }

    database.DB.Delete(user)
    return c.Redirect("/admin/users")
}
