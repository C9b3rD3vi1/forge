package routes

import (
    "github.com/C9b3rD3vi1/forge/handlers"
    "github.com/gofiber/fiber/v2"
)

func SetupPublicRoutes(app *fiber.App) {
    // Public services pages
    app.Get("/services", handlers.ServiceList)        // List all services
    app.Get("/service/:slug", handlers.ServiceView)    // Single service view

    // Public Project routes
    app.Get("/projects", handlers.ProjectList)        // List all projects (public)
    app.Get("/projects/:slug", handlers.ProjectView)  // View single project by slug
    
    // Public Post routes
    app.Get("/posts", handlers.PublicPostList)              // List all posts (public)
    app.Get("/posts/:slug", handlers.PublicPostDetail)        // View single post by slug
    
    // Public GitHub Stats route
    app.Get("/github-stats", handlers.GitHubStatsHandler)
    app.Get("/github-user-stats",handlers.GitHubUserStatsHandler)

}