# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Makefile for common development tasks
- Dockerfile for containerization
- docker-compose.yml for local development with MySQL and Redis
- .editorconfig for consistent code formatting
- .golangci.yml for comprehensive linting configuration
- GitHub Actions CI workflow for automated testing and building
- GitHub Actions security scanning workflow
- CONTRIBUTING.md with development guidelines
- SECURITY.md for vulnerability reporting
- CHANGELOG.md for tracking changes
- Improved .gitignore with comprehensive Go project exclusions

### Fixed
- Import alias conflict in `service/sales/sales_order_grid_service.go`
- Package import declarations to avoid `go vet` errors

### Changed
- Enhanced documentation with contribution guidelines
- Improved project structure documentation

## [1.0.1] - 2025-01-XX

### Added
- Initial public release
- REST API for Magento products, categories, and orders
- Echo web server with RESTful routing
- Basic authentication for all endpoints
- GORM ORM for MySQL
- Global product cache for performance
- Flexible product API with EAV attributes flattened
- Redis integration for caching
- Cron job scheduler
- CLI interface for management tasks
- HTML templates with Tailwind CSS
- Performance monitoring headers
- Request registry and global cache
- Comprehensive README documentation

### Features
- Product flat API with cache (~4x performance improvement)
- Category management
- Sales order grid API
- Product image optimization with WebP support
- Scheduled background jobs
- Multi-environment configuration support

[Unreleased]: https://github.com/Genaker/GoGento/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/Genaker/GoGento/releases/tag/v1.0.1
