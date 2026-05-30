package handlers

import (
	"time"

	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/gofiber/fiber/v2"
)

type ViewsOverTime struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TopContent struct {
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Entity    string `json:"entity"`
	ViewCount int    `json:"view_count"`
}

type SummaryStats struct {
	TotalViews int64 `json:"total_views"`
	TodayViews int64 `json:"today_views"`
	WeekViews  int64 `json:"week_views"`
	MonthViews int64 `json:"month_views"`
}

func AnalyticsViews(c *fiber.Ctx) error {
	days := c.QueryInt("days", 30)
	if days < 1 || days > 365 {
		days = 30
	}
	since := time.Now().AddDate(0, 0, -days)

	var results []ViewsOverTime
	database.DB.Model(&models.PageView{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("created_at > ?", since).
		Group("DATE(created_at)").
		Order("date asc").
		Find(&results)

	return c.JSON(results)
}

func AnalyticsTop(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var posts []models.Post
	database.DB.Order("view_count desc").Limit(limit).Find(&posts)

	var projects []models.Projects
	database.DB.Order("view_count desc").Limit(limit).Find(&projects)

	var services []models.Services
	database.DB.Order("view_count desc").Limit(limit).Find(&services)

	var top []TopContent
	for _, p := range posts {
		top = append(top, TopContent{Title: p.Title, Slug: p.Slug, Entity: "post", ViewCount: p.ViewCount})
	}
	for _, p := range projects {
		top = append(top, TopContent{Title: p.Title, Slug: p.Slug, Entity: "project", ViewCount: p.ViewCount})
	}
	for _, s := range services {
		top = append(top, TopContent{Title: s.Title, Slug: s.Slug, Entity: "service", ViewCount: s.ViewCount})
	}

	return c.JSON(top)
}

func AnalyticsSummary(c *fiber.Ctx) error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekAgo := now.AddDate(0, 0, -7)
	monthAgo := now.AddDate(0, 0, -30)

	var summary SummaryStats

	database.DB.Model(&models.Post{}).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&summary.TotalViews)
	var projViews, svcViews int64
	database.DB.Model(&models.Projects{}).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&projViews)
	database.DB.Model(&models.Services{}).
		Select("COALESCE(SUM(view_count), 0)").
		Scan(&svcViews)
	summary.TotalViews += projViews + svcViews

	database.DB.Model(&models.PageView{}).
		Where("created_at > ?", today).
		Select("COALESCE(COUNT(*), 0)").
		Scan(&summary.TodayViews)

	database.DB.Model(&models.PageView{}).
		Where("created_at > ?", weekAgo).
		Select("COALESCE(COUNT(*), 0)").
		Scan(&summary.WeekViews)

	database.DB.Model(&models.PageView{}).
		Where("created_at > ?", monthAgo).
		Select("COALESCE(COUNT(*), 0)").
		Scan(&summary.MonthViews)

	return c.JSON(summary)
}
