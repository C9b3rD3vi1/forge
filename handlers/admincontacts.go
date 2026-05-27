package handlers

import (
    "math"
    "strconv"

    "github.com/C9b3rD3vi1/forge/database"
    "github.com/C9b3rD3vi1/forge/models"
    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)


// List messages
func AdminContactList(c *fiber.Ctx) error {
    page, _ := strconv.Atoi(c.Query("page", "1"))
    if page < 1 {
        page = 1
    }
    search := c.Query("q", "")

    var contacts []models.ContactMessage
    query := database.DB.Model(&models.ContactMessage{})

    if search != "" {
        query = query.Where(
            "name LIKE ? OR email LIKE ? OR subject LIKE ?",
            "%"+search+"%",
            "%"+search+"%",
            "%"+search+"%",
        )
    }

    pageSize := 20
    var total int64
    query.Count(&total)

    query.Order("created_at DESC").
        Limit(pageSize).
        Offset((page - 1) * pageSize).
        Find(&contacts)

    return c.Render("admin/contacts", fiber.Map{
        "Messages": contacts,
        "Page":     page,
        "Pages":    int(math.Ceil(float64(total) / float64(pageSize))),
        "Search":   search,
    })
}



// View single message
func AdminContactView(c *fiber.Ctx) error {
    idStr := c.Params("id")

    // parse UUID
    id, err := uuid.Parse(idStr)
    if err != nil {
        return c.Status(400).SendString("Invalid message ID")
    }

    var contact models.ContactMessage
    if err := database.DB.First(&contact, "id = ?", id).Error; err != nil {
        return c.Status(404).SendString("Message not found")
    }

    // mark as read
    if !contact.IsRead {
        database.DB.Model(&contact).Update("is_read", true)
    }

    return c.Render("admin/contacts/view", fiber.Map{
        "Message": contact,
    })
}



// Delete message
func AdminContactDelete(c *fiber.Ctx) error {
    idStr := c.Params("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        return c.Status(400).SendString("Invalid message ID")
    }

    database.DB.Delete(&models.ContactMessage{}, "id = ?", id)

    return c.Redirect("/admin/contacts")
}



// Mark as read
func AdminContactMarkRead(c *fiber.Ctx) error {
    idStr := c.Params("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        return c.Status(400).SendString("Invalid message ID")
    }

    database.DB.Model(&models.ContactMessage{}).
        Where("id = ?", id).
        Update("is_read", true)

    return c.Redirect("/admin/contacts")
}



// Mark as unread
func AdminContactMarkUnread(c *fiber.Ctx) error {
    idStr := c.Params("id")
    id, err := uuid.Parse(idStr)
    if err != nil {
        return c.Status(400).SendString("Invalid message ID")
    }

    database.DB.Model(&models.ContactMessage{}).
        Where("id = ?", id).
        Update("is_read", false)

    return c.Redirect("/admin/contacts")
}
