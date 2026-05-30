package handlers

import (
	"time"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

type DashboardStats struct {
	TotalPosts        int64
	PublishedPosts    int64
	TotalProjects     int64
	FeaturedProjects  int64
	TotalServices     int64
	TotalMessages     int64
	UnreadMessages    int64
	TotalUsers        int64
	AdminUsers        int64
	TotalViewCount    int64
	PostsThisWeek     int64
	MessagesThisWeek  int64
	CommentsThisWeek  int64
	UsersThisWeek     int64
}

// AdminDashboard renders the admin dashboard with analytics
func AdminDashboard(c *fiber.Ctx) error {
	admin := config.GetCurrentUser(c)
	if admin == nil {
		return c.Redirect("/admin/login")
	}

	if !admin.IsAdmin {
		return c.SendStatus(fiber.StatusForbidden)
	}

	var posts []models.Post
	var projects []models.Projects
	var services []models.Services
	var users []models.User
	var messages []models.ContactMessage

	database.DB.Order("created_at desc").Find(&posts)
	database.DB.Order("created_at desc").Find(&projects)
	database.DB.Order("created_at desc").Find(&users)
	database.DB.Order("created_at desc").Find(&services)
	database.DB.Order("created_at desc").Find(&messages)

	now := time.Now()
	weekAgo := now.Add(-7 * 24 * time.Hour)

	var stats DashboardStats
	database.DB.Model(&models.Post{}).Count(&stats.TotalPosts)
	database.DB.Model(&models.Post{}).Where("published = ?", true).Count(&stats.PublishedPosts)
	database.DB.Model(&models.Projects{}).Count(&stats.TotalProjects)
	database.DB.Model(&models.Projects{}).Where("featured = ?", true).Count(&stats.FeaturedProjects)
	database.DB.Model(&models.Services{}).Count(&stats.TotalServices)
	database.DB.Model(&models.ContactMessage{}).Count(&stats.TotalMessages)
	database.DB.Model(&models.ContactMessage{}).Where("is_read = ?", false).Count(&stats.UnreadMessages)
	database.DB.Model(&models.User{}).Count(&stats.TotalUsers)
	database.DB.Model(&models.User{}).Where("is_admin = ?", true).Count(&stats.AdminUsers)

	database.DB.Model(&models.Post{}).Where("created_at > ?", weekAgo).Count(&stats.PostsThisWeek)
	database.DB.Model(&models.ContactMessage{}).Where("created_at > ?", weekAgo).Count(&stats.MessagesThisWeek)
	database.DB.Model(&models.Comment{}).Where("created_at > ?", weekAgo).Count(&stats.CommentsThisWeek)
	database.DB.Model(&models.User{}).Where("created_at > ?", weekAgo).Count(&stats.UsersThisWeek)

	database.DB.Model(&models.Projects{}).Select("COALESCE(SUM(view_count), 0)").Scan(&stats.TotalViewCount)
	var svcViews int64
	database.DB.Model(&models.Services{}).Select("COALESCE(SUM(view_count), 0)").Scan(&svcViews)
	stats.TotalViewCount += svcViews

	var topProjects []models.Projects
	database.DB.Order("view_count desc").Limit(5).Find(&topProjects)

	var topServices []models.Services
	database.DB.Order("view_count desc").Limit(5).Find(&topServices)

	var recentPageViews []models.PageView
	database.DB.Order("created_at desc").Limit(10).Find(&recentPageViews)

	return c.Render("admin/dashboard", fiber.Map{
		"Title":          "Admin Dashboard",
		"Admin":          admin,
		"Posts":          posts,
		"Projects":       projects,
		"Services":       services,
		"Users":          users,
		"Messages":       messages,
		"Stats":          stats,
		"TopProjects":    topProjects,
		"TopServices":    topServices,
		"RecentViews":    recentPageViews,
	})
}
