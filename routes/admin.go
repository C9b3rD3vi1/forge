package routes

import (
	"github.com/C9b3rD3vi1/forge/auth"
	"github.com/C9b3rD3vi1/forge/handlers"
	"github.com/C9b3rD3vi1/forge/middleware"
	"github.com/gofiber/fiber/v2"
)

// SetupAdminRoutes sets up the admin routes for the application.
func SetupAdminRoutes(app *fiber.App) {
    // --- Public admin routes (NO middleware) ---
    app.Get("/admin/login", auth.AdminLoginForm)     // GET form
    app.Post("/admin/login", auth.AdminAuthHandler)    // POST form
    app.Get("/admin/verify-otp", auth.ShowOTPPage)    // GET 2FA verification
    app.Post("/admin/verify-otp", auth.VerifyOTPHandler) // POST 2FA verification

    // --- Protected admin routes ---
    admin := app.Group("/admin", middleware.RequireAdminAuth)

    admin.Get("/logout", handlers.AdminLogout)
    admin.Get("/dashboard", handlers.AdminDashboard)
    admin.Get("/profile", handlers.AdminProfile)

    // User management
    // Users
    admin.Get("/users", handlers.AdminUserList)
    admin.Get("/users/create", handlers.AdminCreateUser)
    admin.Post("/users/create", handlers.AdminCreateUser)
    admin.Get("/users/:id", handlers.AdminViewUser)
    admin.Get("/users/:id/edit", handlers.AdminUserEditPage)
    admin.Post("/users/:id/edit", handlers.AdminUserEdit)
    admin.Post("/users/:id/delete", handlers.AdminDeleteUser)

    // Posts
    // Posts CRUD
	admin.Get("/posts", handlers.AdminPostList)                // List all posts
	admin.Get("/posts/new", handlers.AdminFetchTags)           // Fetch all tags
	admin.Get("/posts/new", handlers.AdminNewPostForm)         // Show create form
	admin.Post("/posts", handlers.AdminCreatePost)             // Handle create
	admin.Get("/posts/edit/:id", handlers.AdminEditPostsForm)  // Show edit form
	admin.Post("/posts/update/:id", handlers.AdminUpdatePost)  // Handle update
	admin.Post("/posts/delete/:id", handlers.AdminDeletePost)  // Handle delete
	admin.Get("/posts/:slug", handlers.AdminViewPosts)         // View single post (must be after all /posts/* routes)


    // Projects admin routes
    admin.Get("/projects", handlers.AdminProjectList)           // list all
    admin.Get("/projects/new", handlers.AdminNewProjectPage)  // LIST ALL TECH STACK FOR MULTI SELECT
    admin.Get("/projects/new", handlers.AdminNewProjectForm)    // show create form
    admin.Post("/projects/new", handlers.AdminCreateProject)    // handle create
    admin.Get("/projects/view/:slug", handlers.AdminViewProject) // view single project by slug
    admin.Get("/projects/edit/:id", handlers.AdminEditProjectForm) // show edit form
    admin.Post("/projects/edit/:id", handlers.AdminUpdateProject)  // handle update
    admin.Get("/projects/delete/:id", handlers.AdminDeleteProject) // delete


    
    // Services
    // Services Admin Routes
    admin.Get("/services", handlers.AdminServiceList)            // List all services
    admin.Get("/services/new", handlers.AdminNewServicePage)  // LIST ALL TECH STACK FOR MULTI SELECT
    admin.Get("/services/new", handlers.AdminNewServiceForm)    // Show form to create
    admin.Post("/services/new", handlers.AdminCreateServices)   // Handle create
    admin.Get("/services/edit/:id", handlers.AdminEditServiceForm)  // Show edit form
    admin.Post("/services/edit/:id", handlers.AdminUpdateService)   // Handle update
    admin.Get("/services/delete/:id", handlers.AdminDeleteService)  // Delete
    admin.Get("/services/:slug", handlers.AdminViewService)       // View single service
    
    // Tech Stack Routes (admin)
    admin.Get("/techstacks", handlers.AdminTechStackList)
    admin.Get("/techstacks/new", handlers.AdminNewTechStackForm)
    admin.Post("/techstacks/new", handlers.AdminCreateTechStack)
    admin.Get("/techstacks/edit/:id", handlers.AdminEditTechStackForm)
    admin.Post("/techstacks/edit/:id", handlers.AdminUpdateTechStack)
    admin.Get("/techstacks/delete/:id", handlers.AdminDeleteTechStack)

    
    app.Get("/admin/tags",	handlers.AdminListTags)        // list all tags
    app.Post("/admin/tags", handlers.AdminCreateTag)      // create new tag
    app.Post("/admin/tags/delete/:id", handlers.AdminDeleteTag) // delete tag
    
    
    admin.Get("/contacts", handlers.AdminContactList)
    admin.Get("/contacts/:id", handlers.AdminContactView)
    admin.Post("/contacts/:id/delete", handlers.AdminContactDelete)
    admin.Post("/contacts/:id/read", handlers.AdminContactMarkRead)
    admin.Post("/contacts/:id/unread", handlers.AdminContactMarkUnread)
    admin.Post("/contacts/:id/reply", handlers.AdminContactReply)

    // Settings
    admin.Get("/settings", handlers.AdminSettings)
    admin.Post("/settings", handlers.AdminSettingsUpdate)
    admin.Post("/settings/profile", handlers.AdminProfileUpdate)
    admin.Post("/settings/password", handlers.AdminPasswordUpdate)
    admin.Get("/settings/2fa/setup", auth.AdminSetup2FA)
    admin.Get("/settings/2fa/qrcode", auth.AdminQRCode)
    admin.Post("/settings/2fa/confirm", auth.AdminConfirm2FA)
    admin.Post("/settings/2fa/disable", auth.AdminDisable2FA)
}
