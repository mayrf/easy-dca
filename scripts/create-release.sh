#!/bin/bash

# easy-dca Release Script
# Usage: ./scripts/create-release.sh [version]
# Example: ./scripts/create-release.sh 0.1.0

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
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

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Version number is required"
    echo "Usage: $0 <version>"
    echo "Example: $0 0.1.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Invalid version format. Use semantic versioning (e.g., 0.1.0)"
    exit 1
fi

print_status "Creating release v$VERSION for easy-dca"

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    print_error "Not in a git repository"
    exit 1
fi

# Check if we're on the master branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "master" ]; then
    print_warning "You're not on the master branch (currently on $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Release cancelled"
        exit 1
    fi
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    print_warning "You have uncommitted changes"
    git status --short
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Release cancelled"
        exit 1
    fi
fi

# Check if tag already exists
if git tag -l | grep -q "^v$VERSION$"; then
    print_error "Tag v$VERSION already exists"
    exit 1
fi

# Pull latest changes
print_status "Pulling latest changes from remote..."
git pull origin master

# Run tests (if available)
if [ -f "go.mod" ]; then
    print_status "Running tests..."
    if go test ./...; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
fi

# Build the application (if it's a Go project)
if [ -f "go.mod" ]; then
    print_status "Building application..."
    if go build -v ./cmd/easy-dca; then
        print_success "Build successful"
    else
        print_error "Build failed"
        exit 1
    fi
fi

# Create the tag
print_status "Creating tag v$VERSION..."
git tag -a "v$VERSION" -m "Release v$VERSION"

# Push the tag
print_status "Pushing tag to remote..."
git push origin "v$VERSION"

print_success "Release v$VERSION created and pushed successfully!"

# Display next steps
echo
print_status "Next steps:"
echo "1. GitHub Actions will automatically:"
echo "   - Build and push Docker image to ghcr.io/mayrf/easy-dca:$VERSION"
echo "   - Create GitHub release with changelog"
echo "   - Tag the release as v$VERSION"
echo
echo "2. Monitor the release workflow:"
echo "   https://github.com/mayrf/easy-dca/actions"
echo
echo "3. Verify the release:"
echo "   https://github.com/mayrf/easy-dca/releases/tag/v$VERSION"
echo
echo "4. Test the Docker image:"
echo "   docker pull ghcr.io/mayrf/easy-dca:$VERSION"
echo

print_success "Release process initiated! 🚀" 