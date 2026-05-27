package handlers

import (
	"errors"
	"strings"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// --- Posts ---

// List all posts
func AdminPostList(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var posts []models.Post
	if err := database.DB.Order("created_at desc").Find(&posts).Error; err != nil {
		return c.Status(500).SendString("Error fetching posts")
	}

	return c.Render("admin/posts", fiber.Map{
		"Title": "Admin Post List",
		"Admin": admin,
		"Posts": posts,
	})
}

// Show new post form
func AdminNewPostForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	return c.Render("admin/new_post", fiber.Map{
		"Title": "Add New Post",
		"Admin": admin,
	})
}

func AdminCreatePost(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	title := c.FormValue("title")
	content := c.FormValue("content")
	tagNames := c.FormValue("tag")

	imageURL, _ := utils.UploadImage(c, "image")
	slug := utils.UniqueSlug(database.DB, "posts", title)

	var tags []models.Tag
	for _, t := range strings.Split(tagNames, ",") {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}

		var tag models.Tag
		if err := database.DB.Where("name = ?", t).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tag = models.Tag{Name: t}
				database.DB.Create(&tag)
			} else {
				return c.Status(500).SendString("Error fetching tags")
			}
		}
		tags = append(tags, tag)
	}

	post := models.Post{
		Title:    title,
		Content:  content,
		Slug:     slug,
		ImageURL: imageURL,
		Author:   admin.Username,
	}

	// Step 1: Create the Post. This populates the post.ID
	if err := database.DB.Create(&post).Error; err != nil {
		return c.Status(500).SendString("Error saving post")
	}

	// Step 2: Use Association() to create the links in the join table
	if len(tags) > 0 {
		err := database.DB.Model(&post).Association("Tags").Replace(tags)
		if err != nil {
			return c.Status(500).SendString("Error associating tags with post")
		}
	}

	return c.Redirect("/admin/posts")
}



// View single post
func AdminViewPosts(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	slug := c.Params("slug")
	var post models.Post
	if err := database.DB.Preload("Tags").Where("slug = ?", slug).First(&post).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{"Message": "Post not found"})
	}

	return c.Render("admin/view_post", fiber.Map{
		"Title": "View Post",
		"Admin": admin,
		"Post":  post,
	})
}



// Show edit form
func AdminEditPostsForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var post models.Post
	if err := database.DB.Preload("Tags").First(&post, id).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{"Message": "Post not found"})
	}

	return c.Render("admin/edit_post", fiber.Map{
		"Title": "Edit Post",
		"Admin": admin,
		"Post":  post,
	})
}


// Update post
func AdminUpdatePost(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    id := c.Params("id")
    var post models.Post
    if err := database.DB.Preload("Tags").First(&post, id).Error; err != nil {
        return c.Status(404).Render("errors/404", fiber.Map{"Message": "Post not found"})
    }

    title := c.FormValue("title")
    content := c.FormValue("content")
    tagNames := strings.Split(c.FormValue("tags"), ",")

    // Update fields
    post.Title = title
    post.Content = content
    post.Slug = utils.UniqueSlug(database.DB, "posts", title)

    if imageURL, _ := utils.UploadImage(c, "image"); imageURL != "" {
        post.ImageURL = imageURL
    }

    // Handle tags
    var tags []models.Tag
    for _, t := range tagNames {
        t = strings.TrimSpace(t)
        if t == "" {
            continue
        }

        var tag models.Tag
        if err := database.DB.Where("name = ?", t).First(&tag).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                tag = models.Tag{Name: t}
                database.DB.Create(&tag)
            } else {
                return c.Status(500).SendString("Error fetching tags")
            }
        }
        tags = append(tags, tag)
    }

    // Replace old tags with new ones
    if err := database.DB.Model(&post).Association("Tags").Replace(tags); err != nil {
        return c.Status(500).SendString("Error updating post tags")
    }

    // Validation
    if post.Title == "" || post.Slug == "" {
        return c.Render("admin/edit_post", fiber.Map{
            "Post":  post,
            "Error": "Title and Slug are required",
        })
    }

    if err := database.DB.Save(&post).Error; err != nil {
        return c.Status(500).SendString("Error updating post")
    }

    return c.Redirect("/admin/posts")
}


// Delete post
func AdminDeletePost(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	if err := database.DB.Delete(&models.Post{}, id).Error; err != nil {
		return c.Status(500).SendString("Error deleting post")
	}

	return c.Redirect("/admin/posts")
}


// Fetch all tags and render them in the template
func AdminFetchTags(c *fiber.Ctx) error {
	tags := []models.Tag{}
	if err := database.DB.Find(&tags).Error; err != nil {
		return c.Status(500).SendString("Error fetching tags")
	}

	return c.Render("admin/new_post", fiber.Map{
		"AllTags": tags,
	})
}

