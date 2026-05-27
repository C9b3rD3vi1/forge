#!/bin/bash

# Docker Deployment Script for Forge Hub
# Usage: ./deploy.sh [option]
# Options: deploy | restart | logs | status | down | clean

set -e

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_DIR="/home/user/goapps/forge"
CONTAINER_NAME="cyberdev-hub"
COMPOSE_FILE="docker-compose.yml"

# Print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_permissions() {
    if [ "$EUID" -eq 0 ]; then 
        print_warning "Running as root. Consider using a non-root user."
    fi
}

# Deploy function
deploy() {
    print_info "Starting deployment process..."
    
    # Go to project directory
    cd "$PROJECT_DIR" || { print_error "Failed to cd to $PROJECT_DIR"; exit 1; }
    print_info "Changed to directory: $PROJECT_DIR"
    
    # Pull latest changes from GitHub
    print_info "Pulling latest changes from GitHub..."
    git pull origin main || { print_error "Git pull failed"; exit 1; }
    print_success "Git pull completed successfully"
    
    # Create data directory if it doesn't exist
    mkdir -p data
    print_info "Ensured data directory exists"
    
    # Stop and remove old containers
    print_info "Stopping old containers..."
    docker-compose -f "$COMPOSE_FILE" down || true
    print_success "Old containers stopped"
    
    # Rebuild the image
    print_info "Building Docker image..."
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    print_success "Docker image built successfully"
    
    # Start the container
    print_info "Starting containers..."
    docker-compose -f "$COMPOSE_FILE" up -d
    print_success "Containers started successfully"
    
    # Wait for container to be ready
    sleep 5
    
    # Check if container is running
    if docker ps | grep -q "$CONTAINER_NAME"; then
        print_success "Container $CONTAINER_NAME is running"
        
        # Show recent logs
        print_info "Recent logs:"
        docker logs --tail 20 "$CONTAINER_NAME"
    else
        print_error "Container failed to start"
        docker logs "$CONTAINER_NAME"
        exit 1
    fi
    
    print_success "Deployment completed!"
}

# Restart function
restart() {
    print_info "Restarting containers..."
    cd "$PROJECT_DIR" || exit 1
    docker-compose -f "$COMPOSE_FILE" restart
    print_success "Containers restarted"
}

# View logs function
logs() {
    cd "$PROJECT_DIR" || exit 1
    if [ -n "$1" ]; then
        docker logs -f --tail "$1" "$CONTAINER_NAME"
    else
        docker logs -f --tail 50 "$CONTAINER_NAME"
    fi
}

# Status function
status() {
    print_info "Container status:"
    docker ps -a --filter "name=$CONTAINER_NAME"
    
    print_info "\nResource usage:"
    docker stats --no-stream --filter "name=$CONTAINER_NAME"
}

# Stop function
stop() {
    print_info "Stopping containers..."
    cd "$PROJECT_DIR" || exit 1
    docker-compose -f "$COMPOSE_FILE" down
    print_success "Containers stopped"
}

# Clean function (remove everything)
clean() {
    print_warning "This will remove containers, volumes, and images!"
    read -p "Are you sure? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd "$PROJECT_DIR" || exit 1
        docker-compose -f "$COMPOSE_FILE" down -v
        docker rmi forge_forge 2>/dev/null || true
        print_success "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

# Backup database
backup() {
    print_info "Backing up database..."
    
    BACKUP_DIR="$PROJECT_DIR/backups"
    mkdir -p "$BACKUP_DIR"
    
    BACKUP_FILE="$BACKUP_DIR/server_$(date +%Y%m%d_%H%M%S).db"
    
    if docker exec "$CONTAINER_NAME" test -f /app/data/server.db; then
        docker cp "$CONTAINER_NAME:/app/data/server.db" "$BACKUP_FILE"
        print_success "Database backed up to: $BACKUP_FILE"
        
        # Keep only last 10 backups
        ls -t "$BACKUP_DIR"/server_*.db | tail -n +11 | xargs rm -f 2>/dev/null || true
    else
        print_warning "Database file not found in container"
    fi
}

# Restore database
restore() {
    print_info "Available backups:"
    ls -1 "$PROJECT_DIR"/backups/server_*.db 2>/dev/null || { print_error "No backups found"; exit 1; }
    
    echo
    read -p "Enter backup filename to restore: " BACKUP_FILE
    
    if [ -f "$BACKUP_FILE" ]; then
        docker cp "$BACKUP_FILE" "$CONTAINER_NAME:/app/data/server.db"
        docker restart "$CONTAINER_NAME"
        print_success "Database restored and container restarted"
    else
        print_error "Backup file not found: $BACKUP_FILE"
    fi
}

# Show help
show_help() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  deploy    - Full deployment (pull, build, start)"
    echo "  restart   - Restart containers"
    echo "  logs [n]  - Show logs (last n lines, default 50)"
    echo "  status    - Show container status and stats"
    echo "  stop      - Stop containers"
    echo "  clean     - Remove containers and volumes"
    echo "  backup    - Backup database"
    echo "  restore   - Restore database from backup"
    echo "  help      - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 deploy"
    echo "  $0 logs 100"
    echo "  $0 backup"
}

# Main script
check_permissions

case "${1:-help}" in
    deploy)
        deploy
        ;;
    restart)
        restart
        ;;
    logs)
        logs "$2"
        ;;
    status)
        status
        ;;
    stop)
        stop
        ;;
    clean)
        clean
        ;;
    backup)
        backup
        ;;
    restore)
        restore
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac