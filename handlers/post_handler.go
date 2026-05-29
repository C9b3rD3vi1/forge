package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

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

func AdminNewPostForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var allTags []models.Tag
	database.DB.Order("name asc").Find(&allTags)

	return c.Render("admin/new_post", fiber.Map{
		"Title":   "Add New Post",
		"Admin":   admin,
		"AllTags": allTags,
	})
}

func AdminCreatePost(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	title := c.FormValue("title")
	content := c.FormValue("content")
	excerpt := c.FormValue("excerpt")
	category := c.FormValue("category")
	tagNames := c.FormValue("tag")
	featured := c.FormValue("featured") == "on"

	imageURL, _ := utils.UploadImage(c, "image")
	slug := utils.UniqueSlug(database.DB, "posts", title)

	publishedAt := time.Now()
	if pa := c.FormValue("published_at"); pa != "" {
		if parsed, err := time.Parse("2006-01-02", pa); err == nil {
			publishedAt = parsed
		}
	}

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

	readingTime := utils.ComputeReadingTime(content)

	post := models.Post{
		Title:       title,
		Content:     content,
		Excerpt:     excerpt,
		Category:    category,
		Slug:        slug,
		ImageURL:    imageURL,
		Author:      admin.Username,
		Featured:    featured,
		PublishedAt: publishedAt,
		ReadingTime: readingTime,
	}

	if err := database.DB.Create(&post).Error; err != nil {
		return c.Status(500).SendString("Error saving post")
	}

	if len(tags) > 0 {
		err := database.DB.Model(&post).Association("Tags").Replace(tags)
		if err != nil {
			return c.Status(500).SendString("Error associating tags with post")
		}
	}

	return c.Redirect("/admin/posts")
}

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

	rt := post.ReadingTime
	if rt == 0 {
		rt = utils.ComputeReadingTime(post.Content)
	}

	return c.Render("admin/view_post", fiber.Map{
		"Title": "View Post",
		"Admin": admin,
		"Post":  post,
	})
}

func AdminEditPostsForm(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var post models.Post
	if err := database.DB.Where("id = ?", id).Preload("Tags").First(&post).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{"Message": "Post not found"})
	}

	var allTags []models.Tag
	database.DB.Order("name asc").Find(&allTags)

	return c.Render("admin/edit_post", fiber.Map{
		"Title":   "Edit Post",
		"Admin":   admin,
		"Post":    post,
		"AllTags": allTags,
	})
}

func AdminUpdatePost(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	var post models.Post
	if err := database.DB.Where("id = ?", id).Preload("Tags").First(&post).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{"Message": "Post not found"})
	}

	title := c.FormValue("title")
	content := c.FormValue("content")
	excerpt := c.FormValue("excerpt")
	category := c.FormValue("category")
	tagNames := c.FormValue("tag")
	featured := c.FormValue("featured") == "on"

	slug := utils.UniqueSlug(database.DB, "posts", title)

	publishedAt := post.PublishedAt
	if pa := c.FormValue("published_at"); pa != "" {
		if parsed, err := time.Parse("2006-01-02", pa); err == nil {
			publishedAt = parsed
		}
	}

	post.Title = title
	post.Content = content
	post.Excerpt = excerpt
	post.Category = category
	post.Slug = slug
	post.Featured = featured
	post.PublishedAt = publishedAt
	post.ReadingTime = utils.ComputeReadingTime(content)

	if imageURL, _ := utils.UploadImage(c, "image"); imageURL != "" {
		post.ImageURL = imageURL
	}

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

	if err := database.DB.Model(&post).Association("Tags").Replace(tags); err != nil {
		return c.Status(500).SendString("Error updating post tags")
	}

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

func AdminDeletePost(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	id := c.Params("id")
	if err := database.DB.Where("id = ?", id).Delete(&models.Post{}).Error; err != nil {
		return c.Status(500).SendString("Error deleting post")
	}

	return c.Redirect("/admin/posts")
}

func AdminFetchTags(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil || !admin.IsAdmin {
		return c.Redirect("/admin/login")
	}

	var allTags []models.Tag
	database.DB.Order("name asc").Find(&allTags)

	return c.Render("admin/new_post", fiber.Map{
		"AllTags": allTags,
	})
}
