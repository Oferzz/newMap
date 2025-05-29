#!/bin/bash

# newMap Platform Deployment Script
# This script helps deploy the application to various environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="production"
COMPOSE_FILE="docker-compose.prod.yml"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if required tools are installed
check_requirements() {
    print_info "Checking requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    # Check Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_info "All requirements satisfied!"
}

# Function to load environment variables
load_env() {
    if [ -f .env ]; then
        print_info "Loading environment variables from .env file..."
        export $(cat .env | grep -v '^#' | xargs)
    else
        print_warn ".env file not found. Using default values."
        print_warn "Create a .env file from .env.example for custom configuration."
    fi
}

# Function to validate environment variables
validate_env() {
    print_info "Validating environment variables..."
    
    if [ -z "$MAPBOX_API_KEY" ]; then
        print_error "MAPBOX_API_KEY is not set. Please set it in your .env file."
        exit 1
    fi
    
    if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" == "changeme-use-secure-secret" ]; then
        print_warn "JWT_SECRET is using default value. Please set a secure secret for production!"
    fi
    
    if [ -z "$DB_PASSWORD" ] || [ "$DB_PASSWORD" == "changeme" ]; then
        print_warn "DB_PASSWORD is using default value. Please set a secure password for production!"
    fi
    
    if [ -z "$REDIS_PASSWORD" ] || [ "$REDIS_PASSWORD" == "changeme" ]; then
        print_warn "REDIS_PASSWORD is using default value. Please set a secure password for production!"
    fi
}

# Function to build images
build_images() {
    print_info "Building Docker images..."
    docker-compose -f $COMPOSE_FILE build --no-cache
}

# Function to start services
start_services() {
    print_info "Starting services..."
    docker-compose -f $COMPOSE_FILE up -d
    
    print_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check service health
    docker-compose -f $COMPOSE_FILE ps
}

# Function to run database migrations
run_migrations() {
    print_info "Running database migrations..."
    docker-compose -f $COMPOSE_FILE exec api ./server migrate up
}

# Function to stop services
stop_services() {
    print_info "Stopping services..."
    docker-compose -f $COMPOSE_FILE down
}

# Function to view logs
view_logs() {
    SERVICE=$1
    if [ -z "$SERVICE" ]; then
        docker-compose -f $COMPOSE_FILE logs -f
    else
        docker-compose -f $COMPOSE_FILE logs -f $SERVICE
    fi
}

# Function to backup database
backup_database() {
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    BACKUP_FILE="backup_${TIMESTAMP}.sql"
    
    print_info "Creating database backup: $BACKUP_FILE"
    docker-compose -f $COMPOSE_FILE exec -T postgres pg_dump -U ${DB_USER:-newMap_user} ${DB_NAME:-newMap} > $BACKUP_FILE
    print_info "Backup completed: $BACKUP_FILE"
}

# Function to restore database
restore_database() {
    BACKUP_FILE=$1
    if [ -z "$BACKUP_FILE" ]; then
        print_error "Please provide a backup file to restore"
        exit 1
    fi
    
    if [ ! -f "$BACKUP_FILE" ]; then
        print_error "Backup file not found: $BACKUP_FILE"
        exit 1
    fi
    
    print_warn "This will overwrite the current database. Are you sure? (yes/no)"
    read -r response
    if [ "$response" != "yes" ]; then
        print_info "Restore cancelled"
        exit 0
    fi
    
    print_info "Restoring database from: $BACKUP_FILE"
    docker-compose -f $COMPOSE_FILE exec -T postgres psql -U ${DB_USER:-newMap_user} ${DB_NAME:-newMap} < $BACKUP_FILE
    print_info "Restore completed"
}

# Function to show help
show_help() {
    echo "newMap Platform Deployment Script"
    echo ""
    echo "Usage: ./deploy.sh [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build         Build Docker images"
    echo "  start         Start all services"
    echo "  stop          Stop all services"
    echo "  restart       Restart all services"
    echo "  logs [service] View logs (optionally for specific service)"
    echo "  migrate       Run database migrations"
    echo "  backup        Backup database"
    echo "  restore <file> Restore database from backup file"
    echo "  status        Show service status"
    echo "  clean         Stop services and remove volumes"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  -e, --env     Environment (development|production) [default: production]"
    echo ""
    echo "Examples:"
    echo "  ./deploy.sh start"
    echo "  ./deploy.sh logs api"
    echo "  ./deploy.sh backup"
    echo "  ./deploy.sh restore backup_20231120_120000.sql"
}

# Main script logic
main() {
    # Parse command line arguments
    COMMAND=$1
    shift
    
    # Parse options
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                ENVIRONMENT="$2"
                shift 2
                ;;
            *)
                ARGS="$1"
                shift
                ;;
        esac
    done
    
    # Set compose file based on environment
    if [ "$ENVIRONMENT" == "development" ]; then
        COMPOSE_FILE="docker-compose.yml"
    fi
    
    # Execute command
    case $COMMAND in
        build)
            check_requirements
            load_env
            validate_env
            build_images
            ;;
        start)
            check_requirements
            load_env
            validate_env
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            stop_services
            sleep 2
            check_requirements
            load_env
            validate_env
            start_services
            ;;
        logs)
            view_logs $ARGS
            ;;
        migrate)
            run_migrations
            ;;
        backup)
            backup_database
            ;;
        restore)
            restore_database $ARGS
            ;;
        status)
            docker-compose -f $COMPOSE_FILE ps
            ;;
        clean)
            print_warn "This will remove all data. Are you sure? (yes/no)"
            read -r response
            if [ "$response" == "yes" ]; then
                docker-compose -f $COMPOSE_FILE down -v
                print_info "All services stopped and volumes removed"
            fi
            ;;
        help|"")
            show_help
            ;;
        *)
            print_error "Unknown command: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"