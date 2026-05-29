package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/C9b3rD3vi1/forge/auth"
	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/handlers"
	"github.com/C9b3rD3vi1/forge/routes"
	"github.com/C9b3rD3vi1/forge/utils"

	"github.com/C9b3rD3vi1/forge/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/znbang/gofiber-layout/html"
)

// fibre app main function
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, falling back to defaults")
	}

	// load template engine
	engine := html.New("./templates", ".html")

	//debug mode
	engine.Debug(true)

	// autoreload in dev environment
	engine.Reload(true)

	// Config app layouts
	engine.Layout("layouts/base")
	
	
	engine.AddFunc("parseJSON", utils.ParseJSON)
	engine.AddFunc("trim", utils.Trim)
	engine.AddFunc("add", utils.Add)
	engine.AddFunc("split", utils.SplitString)
	engine.AddFunc("seq", utils.Seq)
	engine.AddFunc("colorClass", utils.ColorClass)
	// time add function
	engine.AddFunc("now", func() string {
		return time.Now().Format("2006-01-02 15:04:05")
	})
	
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	
	// inject global template data (footer services, etc.)
	app.Use(middleware.InjectGlobalData())

	// provide layout for different pages
	app.Use(middleware.DynamicLayoutMiddleware(engine))
	
	//load static files
	app.Static("/static", "./static")
	app.Static("/uploads", "./uploads")
	
	
	// initialize session
	config.InitSession()

	// Initialize the database
 // Initialize database
    db, err := database.InitDB()
    if err != nil {
        log.Fatal("Database initialization failed:", err)
    }

    // Create admin user
    if err := database.CreateAdminUser(db); err != nil {
        log.Fatal("Failed to create admin user:", err)
    }
    log.Println("Admin user created/verified")


    
    
    // Health check endpoint
    app.Get("/health", func(c *fiber.Ctx) error {
        sqlDB, err := database.DB.DB()
        if err != nil || sqlDB.Ping() != nil {
            return c.Status(503).JSON(fiber.Map{"status": "unhealthy"})
        }
        return c.JSON(fiber.Map{"status": "ok"})
    })

    // Setup Adminroutes
    routes.SetupAdminRoutes(app) 
    routes.SetupPublicRoutes(app)
    
    
//	app.Get("/admin/verify", auth.ShowOTPPage)
	//app.Post("/admin/verify", auth.ShowOTPPage)



	// Route to render index.html
	app.Get("/", handlers.HomePageHandler)

	//User registration route
	app.Get("/register", handlers.UserRegisterHandlerForm)
	app.Post("/register", auth.UserRegisterHandler)

	// Route to handle login
	app.Get("/login", handlers.UserLoginHandlerForm)
	// handle post request to login
	app.Post("/login", auth.UserLoginHandler)

	// contact
	app.Get("/contact", handlers.UserContactHandlerForm)
	app.Post("/contact", handlers.UserContactHandler)

	// about us page
	app.Get("/about", handlers.AboutUsHandler)

	// Route to handle logout
	app.Get("/logout", auth.UserLogoutHandler)


	// github stats
	app.Get("/api/github-stats", handlers.GitHubStatsHandler)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3031"
	}
	fmt.Println("Server is running on port " + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
