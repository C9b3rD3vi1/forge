package handlers

import (
	"bytes"
	"html/template"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yuin/goldmark"
	"gorm.io/gorm"
)

type PostListItem struct {
	models.Post
	Excerpt string
}

func PublicPostList(c *fiber.Ctx) error {
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit

	var posts []models.Post
	var totalPosts int64

	database.DB.Model(&models.Post{}).Where("published = ?", true).Count(&totalPosts)

	if err := database.DB.Preload("Tags").Where("published = ?", true).Order("published_at desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		return c.Status(500).SendString("Error fetching posts")
	}

	postsWithExcerpts := make([]PostListItem, len(posts))
	for i, p := range posts {
		excerpt := p.Excerpt
		if excerpt == "" {
			excerpt = p.Content
			if len(excerpt) > 200 {
				excerpt = excerpt[:200] + "..."
			}
		}
		rt := p.ReadingTime
		if rt == 0 {
			rt = utils.ComputeReadingTime(p.Content)
		}
		p.ReadingTime = rt
		postsWithExcerpts[i] = PostListItem{
			Post:    p,
			Excerpt: excerpt,
		}
	}

	totalPages := int(math.Ceil(float64(totalPosts) / float64(limit)))
	pageRange := make([]int, totalPages)
	for i := 1; i <= totalPages; i++ {
		pageRange[i-1] = i
	}

	return c.Render("pages/posts", fiber.Map{
		"Title":       "Blog Posts | Forge.Hub",
		"Posts":       postsWithExcerpts,
		"TotalPages":  totalPages,
		"CurrentPage": page,
		"PageRange":   pageRange,
	})
}

func PublicPostDetail(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var post models.Post

	if err := database.DB.Preload("Tags").Where("slug = ? AND published = ?", slug, true).First(&post).Error; err != nil {
		log.Printf("Post not found: %v", err)
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Post not found",
		})
	}

	rt := post.ReadingTime
	if rt == 0 {
		rt = utils.ComputeReadingTime(post.Content)
	}
	post.ReadingTime = rt

	var buf bytes.Buffer
	md := goldmark.New()
	if err := md.Convert([]byte(post.Content), &buf); err != nil {
		log.Printf("Markdown conversion error for post '%s': %v", post.Title, err)
		return c.Status(500).Render("errors/500", fiber.Map{
			"Message": "Error rendering post content",
		})
	}

	var relatedPosts []models.Post
	if len(post.Tags) > 0 {
		tagIDs := make([]string, len(post.Tags))
		for i, t := range post.Tags {
			tagIDs[i] = t.ID.String()
		}

		if err := database.DB.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id IN ?", tagIDs).
			Where("posts.id != ?", post.ID.String()).
			Where("posts.published = ?", true).
			Preload("Tags").
			Group("posts.id").
			Limit(5).
			Find(&relatedPosts).Error; err != nil {
			log.Printf("Error fetching related posts: %v", err)
			relatedPosts = []models.Post{}
		}
	}

	var comments []models.Comment
	database.DB.Preload("User").Where("post_id = ?", post.ID.String()).Order("created_at asc").Find(&comments)
	if comments == nil {
		comments = []models.Comment{}
	}

	database.DB.Model(&post).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	canonicalURL := post.CanonicalURL
	if canonicalURL == "" {
		canonicalURL = "https://forgehub.tech/posts/" + post.Slug
	}

	return c.Render("pages/postdetail", fiber.Map{
		"Title":       post.Title + " | Forge.Hub",
		"Post":        post,
		"ContentHTML": template.HTML(buf.String()),
		"RelatedPosts": relatedPosts,
		"Comments":    comments,
		"CanonicalURL": canonicalURL,
	})
}

func PublicPostComment(c *fiber.Ctx) error {
	slug := c.Params("slug")
	var post models.Post
	if err := database.DB.Where("slug = ? AND published = ?", slug, true).First(&post).Error; err != nil {
		return c.Status(404).Render("errors/404", fiber.Map{
			"Message": "Post not found",
		})
	}

	sess, err := config.Store.Get(c)
	if err != nil {
		return c.Status(500).SendString("Session error")
	}

	userID, ok := sess.Get("user_id").(string)
	if !ok || userID == "" {
		return c.Redirect("/login?redirect=/posts/" + slug)
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return c.Redirect("/login?redirect=/posts/" + slug)
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", uid).Error; err != nil {
		return c.Redirect("/login?redirect=/posts/" + slug)
	}

	content := strings.TrimSpace(c.FormValue("content"))
	if content == "" {
		return c.Redirect("/posts/" + slug + "#comments")
	}

	comment := models.Comment{
		Content: content,
		UserID:  user.ID,
		PostID:  post.ID,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		log.Printf("Error saving comment: %v", err)
	}

	return c.Redirect("/posts/" + slug + "#comments")
}
