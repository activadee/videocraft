# Development Guidelines

This document outlines development standards and practices for contributing to VideoCraft.

## Code Organization

1. Package Structure: Follow Go standard package layout
2. Interface Design: Define interfaces in consuming packages
3. Dependency Injection: Use constructor injection for services
4. Error Handling: Use wrapped errors with context
5. Testing: Unit tests for business logic, integration tests for workflows

## Git Workflow

1. Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`
2. Create feature branches from main
3. Require code review for all changes
4. Run tests and linting before merge
5. Use semantic versioning for releases