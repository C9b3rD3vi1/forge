package handlers

import (
	"encoding/json"
	"time"
	"fmt"
	"mime/multipart"
	"strings"
	"gorm.io/gorm"

	"github.com/C9b3rD3vi1/forge/config"
	"github.com/C9b3rD3vi1/forge/database"
	"github.com/C9b3rD3vi1/forge/models"
	"github.com/C9b3rD3vi1/forge/utils"
	"github.com/gofiber/fiber/v2"
)

// --- Projects ---

// Show new project form
func AdminNewProjectForm(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    return c.Render("admin/new_project", fiber.Map{
        "Title": "Add New Project",
        "Admin": admin,
    })
}


func AdminNewProjectPage(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    var techStacks []models.TechStack
    if err := database.DB.Order("created_at desc").Find(&techStacks).Error; err != nil {
        return c.Status(500).SendString("Error fetching tech stacks")
    }

    return c.Render("admin/new_project", fiber.Map{
        "Admin":      admin,
        "TechStacks": techStacks, // must match {{ .TechStacks }} in template
    })
}


// Handle project creation with enhanced struct
func AdminCreateProject(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    // Parse form values
    project := models.Projects{
        Title:            c.FormValue("title"),
        Description:      c.FormValue("description"),
        LongDescription:  c.FormValue("long_description"),
        ProblemStatement: c.FormValue("problem_statement"),
        SolutionApproach: c.FormValue("solution_approach"),
        KeyFeatures:      c.FormValue("key_features"),
        ResultsOutcome:   c.FormValue("results_outcome"),

        Link:       c.FormValue("link"),
        GithubLink: c.FormValue("github_link"),
        DemoLink:   c.FormValue("demo_link"),
        DocsLink:   c.FormValue("docs_link"),

        Category:    c.FormValue("category"),
        Difficulty:  c.FormValue("difficulty"),
        ProjectType: c.FormValue("project_type"),
        Tags:        c.FormValue("tags"),

        DevelopmentTime: c.FormValue("development_time"),
        TeamSize:       utils.ParseInt(c.FormValue("team_size")),
        LinesOfCode:    c.FormValue("lines_of_code"),
        Uptime:         c.FormValue("uptime"),
        ResponseTime:   c.FormValue("response_time"),
        UsersCount:     c.FormValue("users_count"),

        Featured:  c.FormValue("featured") == "on",
        Published: c.FormValue("published") == "on",
        Status:    c.FormValue("status"),
    }

    // Parse dates
    if completionDate := c.FormValue("completion_date"); completionDate != "" {
        if parsedDate, err := time.Parse("2006-01-02", completionDate); err == nil {
            project.CompletionDate = &parsedDate
        }
    }

    if startDate := c.FormValue("started_at"); startDate != "" {
        if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
            project.StartedAt = &parsedDate
        }
    }

    // TechStacks handling
    stackIDs := c.FormValue("techstacks")
    if stackIDs != "" {
        ids := strings.Split(stackIDs, ",")
        var techStacks []models.TechStack
        database.DB.Where("id IN ?", ids).Find(&techStacks)
        project.TechStacks = techStacks
    }

    // Upload main image
    if imageURL, err := utils.UploadImage(c, "image"); err == nil && imageURL != "" {
        project.ImageURL = imageURL
    }

    // ---------------------------
    // FIXED: Upload multiple images from "gallery"
    // ---------------------------

    form, err := c.MultipartForm()
    if err == nil && form.File != nil {
        galleryFiles := form.File["gallery"]
        var galleryURLs []string
    
        for idx, file := range galleryFiles {
            tempField := fmt.Sprintf("gallery_%d", idx)
    
            // Insert file into the form under the temporary name
            form.File[tempField] = []*multipart.FileHeader{file}
    
            // Call UploadImage normally
            galleryURL, err := utils.UploadImage(c, tempField)
            if err == nil {
                galleryURLs = append(galleryURLs, galleryURL)
            }
    
            // Remove temp field
            delete(form.File, tempField)
        }
    
        if len(galleryURLs) > 0 {
            if galleryJSON, err := json.Marshal(galleryURLs); err == nil {
                project.Gallery = string(galleryJSON)
            }
        }
    }

    // Unique slug
    project.Slug = utils.UniqueSlug(database.DB, "projects", project.Title)

    // Save to DB
    if err := database.DB.Create(&project).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Error saving project: " + err.Error())
    }

    return c.Redirect("/admin/projects")
}


// List all projects
func AdminProjectList(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    var projects []models.Projects
    // Use Preload("TechStacks") to eager load the many-to-many relationship
    result := database.DB.Preload("TechStacks").Order("created_at desc").Find(&projects)
    
    if result.Error != nil {
        // Handle the error appropriately, e.g., log it and return an error page.
        // For now, let's just show a simple error.
        return c.Status(500).SendString("Error loading projects")
    }

    return c.Render("admin/projects", fiber.Map{
        "Title":    "Manage Projects",
        "Admin":    admin,
        "Projects": projects,
    })
}


// View single project
func AdminViewProject(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    slug := c.Params("slug")
    var project models.Projects
    if err := database.DB.Where("slug = ?", slug).First(&project).Error; err != nil {
        return c.Status(404).Render("errors/404", fiber.Map{"Message": "Project not found"})
    }

    return c.Render("admin/view_project", fiber.Map{
        "Title":   "View Project",
        "Admin":   admin,
        "Project": project,
    })
}


// Show edit form with tech stacks
func AdminEditProjectForm(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    id := c.Params("id")
    var project models.Projects
    
    // Preload TechStacks to show in form
    if err := database.DB.Where("id = ?", id).Preload("TechStacks").First(&project).Error; err != nil {
        return c.Status(404).Render("errors/404", fiber.Map{"Message": "Project not found"})
    }

    // Get all available tech stacks for dropdown
    var allTechStacks []models.TechStack
    database.DB.Find(&allTechStacks)

    return c.Render("admin/edit_project", fiber.Map{
        "Title":        "Edit Project",
        "Admin":        admin,
        "Project":      project,
        "TechStacks":   allTechStacks,
    })
}

// Handle project update with enhanced struct
func AdminUpdateProject(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

    id := c.Params("id")
    var project models.Projects
    if err := database.DB.Where("id = ?", id).First(&project).Error; err != nil {
        return c.Status(404).Render("errors/404", fiber.Map{"Message": "Project not found"})
    }

    // Update ALL fields consistently with create handler
    project.Title = c.FormValue("title")
    project.Description = c.FormValue("description")
    project.LongDescription = c.FormValue("long_description")
    project.ProblemStatement = c.FormValue("problem_statement")
    project.SolutionApproach = c.FormValue("solution_approach")
    project.KeyFeatures = c.FormValue("key_features")
    project.ResultsOutcome = c.FormValue("results_outcome")
    
    // Links
    project.Link = c.FormValue("link")
    project.GithubLink = c.FormValue("github_link")
    project.DemoLink = c.FormValue("demo_link")
    project.DocsLink = c.FormValue("docs_link")
    
    // Categorization
    project.Category = c.FormValue("category")
    project.Difficulty = c.FormValue("difficulty")
    project.ProjectType = c.FormValue("project_type")
    project.Tags = c.FormValue("tags")
    
    // Stats & Metrics
    project.DevelopmentTime = c.FormValue("development_time")
    project.LinesOfCode = c.FormValue("lines_of_code")
    project.Uptime = c.FormValue("uptime")
    project.ResponseTime = c.FormValue("response_time")
    project.UsersCount = c.FormValue("users_count")
    
    // Team size needs parsing
    if teamSizeStr := c.FormValue("team_size"); teamSizeStr != "" {
        project.TeamSize = utils.ParseInt(teamSizeStr)
    }
    
    // Status
    project.Featured = c.FormValue("featured") == "on"
    project.Published = c.FormValue("published") == "on"
    project.Status = c.FormValue("status")
    
    // Parse dates
    if completionDate := c.FormValue("completion_date"); completionDate != "" {
        if parsedDate, err := time.Parse("2006-01-02", completionDate); err == nil {
            project.CompletionDate = &parsedDate
        } else {
            project.CompletionDate = nil // Clear if invalid
        }
    }
    
    if startDate := c.FormValue("started_at"); startDate != "" {
        if parsedDate, err := time.Parse("2006-01-02", startDate); err == nil {
            project.StartedAt = &parsedDate
        } else {
            project.StartedAt = nil // Clear if invalid
        }
    }
    
    // TechStacks - Need to handle relationship update
    stackIDs := c.FormValue("techstacks")
    if stackIDs != "" {
        ids := strings.Split(stackIDs, ",")
        var techStacks []models.TechStack
        if len(ids) > 0 {
            database.DB.Where("id IN ?", ids).Find(&techStacks)
            
            // Clear existing associations and set new ones
            database.DB.Model(&project).Association("TechStacks").Clear()
            project.TechStacks = techStacks
        }
    } else {
        // Clear tech stacks if none selected
        database.DB.Model(&project).Association("TechStacks").Clear()
        project.TechStacks = nil
    }
    
    // Update main image if provided
    if imageURL, err := utils.UploadImage(c, "image"); err == nil && imageURL != "" {
        project.ImageURL = imageURL
    }
    
    // Handle gallery updates
    form, err := c.MultipartForm()
    if err == nil && form.File != nil {
        galleryFiles := form.File["gallery"]
        if len(galleryFiles) > 0 {
            // Process new gallery images
            var galleryURLs []string
            
            // Parse existing gallery if any
            if project.Gallery != "" {
                json.Unmarshal([]byte(project.Gallery), &galleryURLs)
            }
            
            // Add new images
            for idx, file := range galleryFiles {
                tempField := fmt.Sprintf("gallery_%d", idx)
                form.File[tempField] = []*multipart.FileHeader{file}
                
                galleryURL, err := utils.UploadImage(c, tempField)
                if err == nil {
                    galleryURLs = append(galleryURLs, galleryURL)
                }
                
                delete(form.File, tempField)
            }
            
            // Save updated gallery
            if galleryJSON, err := json.Marshal(galleryURLs); err == nil {
                project.Gallery = string(galleryJSON)
            }
        }
    }

    // Save all changes to database
    if err := database.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(&project).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).SendString("Error updating project: " + err.Error())
    }

    return c.Redirect("/admin/projects")
}

// Delete project
func AdminDeleteProject(c *fiber.Ctx) error {
    admin := config.GetCurrentUser(c)
    if admin == nil || !admin.IsAdmin {
        return c.Redirect("/admin/login")
    }

	id := c.Params("id")
	if err := database.DB.Where("id = ?", id).Delete(&models.Projects{}).Error; err != nil {
		return c.Status(500).SendString("Error deleting project")
	}

    return c.Redirect("/admin/projects")
}
