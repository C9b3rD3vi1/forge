package handlers

import (
	"bytes"
	"html/template"
	"log"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
	"github.com/yuin/goldmark"
)

// PostWithExcerpt is a helper struct to send excerpts to templates
type PostWithExcerpt struct {
	models.Post
	Excerpt string
}

// PublicPostList lists all posts with excerpts
func PublicPostList(c *fiber.Ctx) error {
	var posts []models.Post

	if err := database.DB.Preload("Tags").Order("created_at desc").Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		return c.Status(500).SendString("Error fetching posts")
	}

	// Generate excerpts (first 150 chars)
	postsWithExcerpts := make([]PostWithExcerpt, len(posts))
	for i, p := range posts {
		excerpt := p.Content
		if len(excerpt) > 150 {
			excerpt = excerpt[:150] + "..."
		}
		postsWithExcerpts[i] = PostWithExcerpt{
			Post:    p,
			Excerpt: excerpt,
		}
	}

	return c.Render("pages/posts", fiber.Map{
		"Posts": postsWithExcerpts,
	})
}

// PublicPostDetail shows a single post with safe HTML and related posts
func PublicPostDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var post models.Post

	log.Printf("Fetching post with slug: %s", slug)
	if err := database.DB.Preload("Tags").Where("slug = ?", slug).First(&post).Error; err != nil {
		log.Printf("Post not found: %v", err)
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Post not found",
		})
	}

	// Convert Markdown to HTML
	var buf bytes.Buffer
	md := goldmark.New()
	if err := md.Convert([]byte(post.Content), &buf); err != nil {
		log.Printf("Markdown conversion error for post '%s': %v", post.Title, err)
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Error rendering post content",
		})
	}

	// Fetch related posts (at least one matching tag)
	var relatedPosts []models.Post
	if len(post.Tags) > 0 {
		tagIDs := make([]string, len(post.Tags))
		for i, t := range post.Tags {
			tagIDs[i] = t.ID.String() // assuming UUID stored as string
		}

		if err := database.DB.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id IN ?", tagIDs).
			Where("posts.id != ?", post.ID.String()).
			Preload("Tags").
			Group("posts.id").
			Limit(5).
			Find(&relatedPosts).Error; err != nil {
			log.Printf("Error fetching related posts: %v", err)
			relatedPosts = []models.Post{}
		}
	}

	return c.Render("pages/postdetail", fiber.Map{
		"Post":         post,
		"ContentHTML":  template.HTML(buf.String()), // safe HTML
		"RelatedPosts": relatedPosts,
	})
}
