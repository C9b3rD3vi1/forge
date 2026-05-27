package handlers

import (
	"fmt"
	"html/template"
	"math"
	"strconv"
	"path/filepath"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	//"github.com/google/uuid"
)

func BlogHandler(c *fiber.Ctx) error {
	//get page querry parameter
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	// define pagination limit
	limit := 10
	offset := (page - 1) * limit

	// Get posts from database
	var blogPosts []models.BlogPost
	var totalPosts int64

	//count blogposts created in the database
	// Count total blog posts
	if err := database.DB.Model(&models.BlogPost{}).Count(&totalPosts).Error; err != nil {
		return c.Status(500).SendString("Failed to count posts")
	}

	// Fetch pagination of blog posts
	if err := database.DB.Model(&models.BlogPost{}).Limit(limit).Offset(offset).Find(&blogPosts).Error; err != nil {
		return c.Status(500).SendString("Failed to fetch posts")
	}

	// calculate total pages
	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))

	// calculate current page
	currentPage := page
	if currentPage > totalPages {
		currentPage = totalPages
	}

	// calculate page range
	pageRange := make([]int, totalPages)
	for i := 1; i <= totalPages; i++ {
		pageRange[i-1] = i
	}

	//	posts := models.CreateBlog() // ← call dummy data function here

	return c.Render("pages/blog", fiber.Map{
		"Posts":       blogPosts, // []any{}, //or[]models.Post{}
		"TotalPages":  totalPages,
		"CurrentPage": currentPage,
		"PageRange":   pageRange,
	})
}

func BlogPostHandler(c *fiber.Ctx) error {
	//slug := c.Params("slug")
	//post := models.CreateSamplePost() // or fetch from DB

	return c.Render("pages/post", fiber.Map{
		//"post": post,
	})
}

func ShowCreateBlogForm(c *fiber.Ctx) error {
	return c.Render("admin/create_blog", fiber.Map{})
}

func CreateBlogPostHandler(c *fiber.Ctx) error {
	title := c.FormValue("title")
	slug := c.FormValue("slug")
	excerpt := c.FormValue("excerpt")
	content := c.FormValue("content")
	author := c.FormValue("author")
	imageURL := ""

	if title == "" || slug == "" || excerpt == "" {
		return c.Render("admin/create_blog", fiber.Map{
			"Error": "Title, Slug, and Excerpt are required",
		})
	}

	var exists models.BlogPost
	if err := database.DB.Where("slug = ?", slug).First(&exists).Error; err == nil {
		return c.Render("admin/create_blog", fiber.Map{
			"Error": "Slug already exists. Choose a different one.",
		})
	}

	// Handle image upload
	file, err := c.FormFile("image")
	if err == nil {
		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s%s", utils.UUID(), ext)
		path := fmt.Sprintf("./uploads/%s", filename)
		if err := c.SaveFile(file, path); err != nil {
			return c.Render("admin/create_blog", fiber.Map{"Error": "Failed to upload image."})
		}
		imageURL = "/uploads/" + filename
	}

	blogPost := models.BlogPost{
		Title:    title,
		Slug:     slug,
		Excerpt:  excerpt,
		Content:  content,
		ImageURL: imageURL,
		Author:   author,
	}

	if err := database.DB.Create(&blogPost).Error; err != nil {
		return c.Render("admin/create_blog", fiber.Map{
			"Error": "Failed to create blog post",
		})
	}

	return c.Redirect("/blog")
}


func BlogDetailsHandler(c *fiber.Ctx) error {
	slug := c.Params("slug")

	var post models.BlogPost
	if err := database.DB.Where("slug = ?", slug).First(&post).Error; err != nil {
		return c.Status(404).Render("errrs/404", fiber.Map{
			"Message": "Blog post not found",
		})
	}

	// Mark content as safe HTML
	type SafePost struct {
		models.BlogPost
		SafeContent template.HTML
	}

	safePost := SafePost{
		BlogPost:    post,
		SafeContent: template.HTML(post.Content),
	}

	return c.Render("pages/blog_detail", fiber.Map{
		"Post": safePost,
	})
}


