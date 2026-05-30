# 🛠️ Forge Hub

Forge Hub is a high-performance, production-ready professional portfolio and blogging platform designed for software engineers and creators. It provides a comprehensive suite of tools to showcase projects, publish technical blog posts, manage service offerings with tech stack associations, and handle client inquiries—all managed through a robust administrative dashboard.

![Project Preview](./static/images/main.png)

## ✨ Key Features

### 🚀 Project Showcase
- **Deep Case Studies**: Beyond simple links, showcase projects with problem statements, solution approaches, key features, and measurable results.
- **Technical Metrics**: Track and display project performance metrics like uptime, response time, and scale.
- **Rich Metadata**: Categorize projects by difficulty, type (Open Source, Enterprise, etc.), and associate them with a dynamic tech stack.
- **SEO Ready**: Per-project meta descriptions, canonical URLs, and auto-tracked view counts.

### ✍️ Professional Blogging
- **Markdown Support**: Write technical articles using a clean Markdown-based system with syntax highlighting.
- **Taxonomy**: Organize content using categories and tags for better discoverability.
- **SEO Metadata**: Per-post meta descriptions, canonical URLs, Open Graph, Twitter Card, and JSON-LD structured data.

### 🛠️ Service Management
- **Service Catalog**: Create and manage service offerings with full descriptions, tech stack associations, and galleries.
- **Frontend Display**: Public services page with filtering, pagination, and individual service detail views.
- **Contact Integration**: Visitors can select relevant services when submitting the contact form.

### 🛡️ Admin Powerhouse
- **Full CMS**: Complete CRUD operations for Projects, Blog Posts, Services, Tech Stacks, and Tags.
- **Contact Management**: Inbox with read/unread status, search, pagination, and an integrated reply system.
- **User & Settings Management**: Manage user accounts, site configuration, and 2FA security.
- **Secure Access**: Protected by password authentication with TOTP two-factor authentication.

### 🌐 Integrations & Utilities
- **GitHub Stats**: Real-time integration of GitHub contribution graphs and repository stats.
- **Email System**: Auto-reply to contact form submissions, admin notifications for new messages, and admin reply via the dashboard.
- **Tech Stack Taxonomy**: Shared tech stack model associated with both projects and services.
- **Health Monitoring**: Built-in `/health` endpoint for container orchestration and uptime monitoring.

## 🏗️ Tech Stack

| Layer | Technology |
| :--- | :--- |
| **Language** | [Go (Golang) 1.23+](https://go.dev/) |
| **Web Framework** | [Fiber v2](https://gofiber.io/) |
| **Database** | [SQLite](https://www.sqlite.org/) via [GORM](https://gorm.io/) |
| **Session** | Cookie-based (fiber session store) |
| **Templating** | Go HTML/template with Fiber layout engine |
| **Frontend** | HTML5, CSS3, vanilla JavaScript |
| **Deployment** | Docker, Docker Compose |
| **Security** | bcrypt, TOTP 2FA (via pquerna/otp), UUID |

## 🚀 Getting Started

### Prerequisites
- Go 1.23+ (for local development)
- Docker & Docker Compose (for production deployment)

### Local Installation
1. **Clone the repository**
   ```bash
   git clone https://github.com/C9b3rD3vi1/forge.git
   cd forge
   ```

2. **Environment Setup**
   Copy `.env.example` to `.env` and fill in your values:
   ```bash
   cp .env.example .env
   ```
   ```env
   APP_PORT=3031
   DB_PATH=server.db
   SESSION_SECRET=change-me-to-a-random-string
   ADMIN_USERNAME=admin
   ADMIN_EMAIL=admin@example.com
   ADMIN_PASSWORD=admin123
   GITHUB_USERNAME=your-username
   SMTP_HOST=smtp.gmail.com
   SMTP_PORT=587
   SMTP_USER=your@email.com
   SMTP_PASSWORD=your-app-password
   TZ=Africa/Nairobi
   ```

3. **Run the application**
   ```bash
   go run main.go
   ```
   The server will start on `http://localhost:3031`.

### Docker Deployment (Recommended)
Forge Hub comes with a fully containerized setup.

1. **Build and Run**
   ```bash
   docker-compose up -d --build
   ```

2. **Verify Health**
   ```bash
   curl http://localhost:3031/health
   ```

## 🚢 Production Management

The project includes a comprehensive lifecycle management script `deploy.sh` to handle updates and maintenance.

### Deployment Lifecycle
```bash
chmod +x deploy.sh

# Full deployment (Pull latest, rebuild, and restart)
./deploy.sh deploy

# Restart the service
./deploy.sh restart

# Check container status and resource usage
./deploy.sh status

# View real-time logs
./deploy.sh logs 100
```

### Database Maintenance
Ensure your data is safe with built-in backup and restore capabilities:

```bash
# Backup the SQLite database
./deploy.sh backup

# Restore from a specific backup file
./deploy.sh restore
```

## 📂 Project Architecture

Forge Hub follows a **Modular Monolith Architecture**, designed for high efficiency and low latency.

### 🗺️ System Diagram
```mermaid
graph TB
    subgraph Client_Layer [Client Layer]
        Visitor((Visitor Browser))
        Admin((Admin Browser))
    end

    subgraph Presentation_Layer [Presentation Layer — Fiber v2]
        direction TB
        Router[Fiber Router]
        MW_Inject[InjectGlobalData\nfooter services, login state]
        MW_Layout[DynamicLayoutMiddleware\npublic ↔ admin layout]
        MW_Auth[RequireAdminAuth\nsession check + admin guard]
        TEngine[Go Template Engine\nlayouts/admin.html\nlayouts/base.html]
        Static[Static Files\n/static /uploads]
    end

    subgraph Handler_Layer [Handler Layer]
        direction TB
        PublicH[Public Handlers\nHome · About · Contact\nServices · Projects · Posts]
        AdminH[Admin Handlers\nDashboard · CRUD Services\nCRUD Projects · CRUD Posts\nCRUD TechStacks · CRUD Tags\nContacts · Settings]
        GitHubH[GitHub Handlers\nStats · Chart · User Stats]
        AuthH[Auth Handlers\nLogin · Register · TOTP 2FA]
    end

    subgraph Data_Layer [Persistence]
        direction LR
        DB[(SQLite\nGORM)]
        Session[(Cookie Session\nfiber session store)]
        Uploads[(File System\n/uploads)]
    end

    subgraph External_Layer [External Integrations]
        GH_API[GitHub REST API]
        SMTP[SMTP Server\ncontact auto-reply\nadmin notify · admin reply]
    end

    %% Request flow — Public
    Visitor -->|GET/POST| Router
    Router --> MW_Inject
    MW_Inject --> MW_Layout
    MW_Layout -->|public routes| PublicH
    MW_Layout -->|/health| Router
    PublicH --> DB
    PublicH --> SMTP

    %% Request flow — Admin
    Admin -->|GET/POST /admin/*| Router
    Router --> MW_Inject
    MW_Inject --> MW_Layout
    MW_Layout -->|/admin/login, /admin/verify-otp| AuthH
    MW_Layout -->|/admin/*| MW_Auth
    MW_Auth --> AdminH
    MW_Auth --> AuthH
    AdminH --> DB
    AdminH --> Uploads
    AdminH --> SMTP

    %% Auth
    AuthH --> Session
    AuthH --> DB

    %% GitHub
    GitHubH --> GH_API

    %% Template rendering
    PublicH --> TEngine
    AdminH --> TEngine
    AuthH --> TEngine
    TEngine --> Visitor
    TEngine --> Admin

    %% Static
    Static --> Visitor
    Static --> Admin

    %% Styling
    style Client_Layer fill:#f9f9f9,stroke:#333,stroke-width:2px
    style Presentation_Layer fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    style Handler_Layer fill:#fff3e0,stroke:#e65100,stroke-width:2px
    style Data_Layer fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    style External_Layer fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    
    style DB fill:#C5E1A5,stroke:#33691E
    style Session fill:#FFCDD2,stroke:#B71C1C
    style GH_API fill:#D1C4E9,stroke:#311B92
    style SMTP fill:#BBDEFB,stroke:#0D47A1
```

### ⚙️ Architectural Analysis

#### Request Lifecycle
1. **Entry**: HTTP request → Fiber Server.
2. **Middleware Chain**: `InjectGlobalData` (footer services, login state) → `DynamicLayoutMiddleware` (public/admin layout switch) → `RequireAdminAuth` (admin routes only).
3. **Routing**: Dispatch to specialized Handlers (Public, Admin, Auth, GitHub).
4. **Persistence**: Business logic → GORM → SQLite. Image/file uploads written to `/uploads`.
5. **Response**: Template Engine → HTML → Client. Static assets (`/static`, `/uploads`) served directly.

#### Key Technical Decisions
- **SQLite**: Simple, zero-config database — no external DB server required.
- **Server-Side Rendering**: Pure Go `html/template` rendering for maximum SEO and performance.
- **Defense-in-Depth**: bcrypt password hashing + TOTP 2FA for administrative security.
- **Async Email**: Email sends are non-blocking goroutines — the HTTP response is never blocked by SMTP.

### 🎨 Component Legend
| Color | Layer | Responsibility |
| :--- | :--- | :--- |
| **Blue** | Presentation | HTTP Routing, Middleware Chain, Static Files |
| **Orange** | Handler | Business Logic — Public, Admin, Auth, GitHub Handlers |
| **Green** | Data | SQLite Persistence & Cookie Sessions & File Uploads |
| **Purple** | External | Third-party API Integrations & SMTP Email |

### 📂 Directory Structure
```text
.
├── auth/           # Authentication logic (login, TOTP 2FA, session management)
├── config/         # Session store, Redis connection test
├── database/       # DB initialization, auto-migration, seed data
├── handlers/       # Business logic for Public, Admin, and API routes
├── middleware/     # Auth guards, layout injection, global data middleware
├── models/         # GORM data models (Project, Post, Service, User, Settings, TechStack)
├── routes/         # Public and Admin route definitions
├── static/         # CSS, favicon, images, logos, site.webmanifest
├── templates/      # HTML templates (Admin panels, Public pages, Email templates)
├── uploads/        # User-uploaded images (icons, project/service images)
└── utils/          # Helper functions (SMTP email, image upload, GitHub API, HTML escaping)
```


## 🛡️ Security Note
Forge Hub implements:
- **Secure Password Storage**: Using `bcrypt` for hashing.
- **Session Isolation**: Managed via Redis.
- **Two-Factor Authentication**: Admin access requires TOTP verification.
- **Container Hardening**: Docker configuration includes `no-new-privileges` and limited logging to prevent disk exhaustion.

---
© 2026 Forge Hub. Built with 💙 using Go.
