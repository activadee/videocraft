# Claude Code Action Integration

This document explains the Claude AI integration for automated documentation review in the VideoCraft project.

## Overview

The Claude Code Action automatically reviews documentation changes and ensures code-documentation synchronization when commits are made to the main branch. This helps maintain high-quality, accurate documentation throughout the project lifecycle.

## Features

### 📝 Automatic Documentation Review
- Automatically reviews all documentation changes for clarity, accuracy, and completeness
- Checks technical accuracy of code examples and API documentation
- Ensures consistency with project style and structure
- Validates links and references

### 🔄 Code-Documentation Synchronization
- Detects when code changes might require documentation updates
- Identifies API changes that need documentation updates
- Checks for new configuration options that need documentation
- Highlights breaking changes that require documentation updates

### 🔒 Security Documentation Monitoring
- Special review for security-related commits
- Ensures security best practices are documented
- Validates security configuration examples
- Cross-references with security-first.md

### 🚀 API Documentation Sync
- Monitors API-related changes
- Ensures endpoint documentation stays current
- Validates request/response examples
- Checks authentication and error documentation

## Configuration

### Environment Variables Required

Add these secrets to your GitHub repository:

```bash
# Required for Claude API access
ANTHROPIC_API_KEY=your_anthropic_api_key_here

# GitHub token (automatically provided)
GITHUB_TOKEN=${{ secrets.GITHUB_TOKEN }}
```

### Setting Up Anthropic API Key

1. **Get Claude API Access**:
   - Sign up at [Anthropic Console](https://console.anthropic.com/)
   - Create an API key
   - Note: A Claude Pro subscription is recommended for optimal performance

2. **Add to GitHub Secrets**:
   ```bash
   # In your repository settings
   Settings > Secrets and variables > Actions > New repository secret
   Name: ANTHROPIC_API_KEY
   Value: your_api_key_here
   ```

## Workflow Triggers

The documentation review workflow triggers on:

### 📄 Documentation Changes
```yaml
paths:
  - '**.md'
  - 'docs/**'
  - 'CLAUDE.md'
  - '**/CLAUDE.md'
  - 'README.md'
  - 'CHANGELOG.md'
```

### 💻 Code Changes
```yaml
paths:
  - 'internal/**/*.go'
  - 'pkg/**/*.go'
  - 'cmd/**/*.go'
  - 'scripts/**/*.py'
```

### 🔍 Special Triggers
- **Security commits**: Commit messages containing "security", "fix", or "vulnerability"
- **API commits**: Commit messages containing "api", "endpoint", or "handler"

## Review Focus Areas

### VideoCraft-Specific Reviews

#### Progressive Subtitles System
- Word-level timing explanations
- Scene-based timing calculations
- ASS subtitle generation documentation
- Python-Go integration details

#### Security Documentation
- Authentication configuration
- SSL/TLS setup instructions
- Container security settings
- Input validation guidelines
- Rate-limiting configuration

#### API Documentation
- Video generation endpoints
- Job management endpoints
- Authentication requirements
- Error handling documentation

## Output and Feedback

### GitHub Comments
The Claude AI will post comments on commits with:
- ✅ Documentation quality assessment
- 📋 Specific improvement suggestions
- 🔗 Broken link identification
- 📚 Missing documentation notifications
- 🔄 Synchronization recommendations

### Issue Creation
For significant issues, the action may create GitHub issues for:
- Missing API documentation
- Outdated security information
- Broken links or references
- Inconsistent terminology
- Missing configuration examples

## Best Practices

### For Contributors

1. **Update Documentation with Code Changes**:
   ```bash
   # When adding new features
   git add internal/api/new_handler.go
   git add internal/api/CLAUDE.md  # Update API docs
   git commit -m "feat: add new video processing endpoint
   
   - Add POST /api/v1/videos/process endpoint
   - Update API documentation with examples
   - Add rate limiting and authentication info"
   ```

2. **Security-Related Changes**:
   ```bash
   # Include security documentation updates
   git add internal/services/auth_service.go
   git add security-first.md  # Update security docs
   git commit -m "security: implement enhanced authentication
   
   - Add JWT token validation
   - Update security documentation
   - Include configuration examples"
   ```

3. **API Changes**:
   ```bash
   # Always update API documentation
   git add internal/api/handlers/video.go
   git add internal/api/CLAUDE.md
   git commit -m "api: enhance video generation parameters
   
   - Add quality and format parameters
   - Update API documentation with examples
   - Include validation rules"
   ```

### Documentation Quality Standards

#### Completeness Checklist
- [ ] All public APIs are documented
- [ ] Configuration options are explained
- [ ] Examples are provided for complex features
- [ ] Security considerations are addressed
- [ ] Deployment instructions are complete

#### Technical Accuracy
- [ ] Code examples compile and run
- [ ] API endpoints match implementation
- [ ] Configuration examples are valid
- [ ] Version information is current
- [ ] Dependencies are correctly listed

#### Consistency Standards
- [ ] Terminology is used consistently
- [ ] Code style matches project standards
- [ ] Documentation structure follows patterns
- [ ] Cross-references are accurate
- [ ] Formatting is consistent

## Troubleshooting

### Common Issues

#### 1. API Key Not Working
```yaml
Error: 401 Unauthorized
```
**Solution**: Check that `ANTHROPIC_API_KEY` is correctly set in repository secrets.

#### 2. Workflow Not Triggering
**Solution**: Ensure file paths match the trigger patterns in `.github/workflows/documentation-review.yml`.

#### 3. Too Many API Calls
**Solution**: The workflow is configured with appropriate triggers to minimize API usage while maintaining quality.

### Debugging

Enable debug mode by adding to the workflow:
```yaml
env:
  ACTIONS_STEP_DEBUG: true
  ACTIONS_RUNNER_DEBUG: true
```

## Configuration Customization

### Modify Review Focus

Edit `.github/claude-config.yml` to customize:

```yaml
# Add custom review areas
custom_focus:
  deployment:
    - "Docker configuration accuracy"
    - "Environment variable documentation"
    - "Scaling instructions"
```

### Adjust Sensitivity

```yaml
# Modify trigger sensitivity
triggers:
  documentation_changes: "always"  # always, major_only, manual
  code_changes: "api_only"         # always, api_only, major_only
  security_changes: "always"       # always, critical_only
```

## Integration with Development Workflow

### Pre-commit Hook Integration
```bash
# .githooks/pre-commit
#!/bin/bash
echo "📝 Reminder: Update documentation for any API or feature changes"
echo "🔍 Claude AI will review documentation on push to main"
```

### Pull Request Template
```markdown
## Documentation Checklist
- [ ] Updated relevant CLAUDE.md files
- [ ] Added examples for new features
- [ ] Updated API documentation if endpoints changed
- [ ] Reviewed security implications and updated docs
- [ ] Verified all links and references work
```

## Performance and Costs

### API Usage Optimization
- Reviews only trigger on relevant file changes
- Batches multiple changes in a single review
- Uses efficient prompts to minimize token usage
- Caches results to avoid duplicate reviews

### Expected Monthly Usage
- Small projects: ~$5-10/month
- Medium projects: ~$15-25/month
- Large projects: ~$30-50/month

*Costs depend on commit frequency and documentation size*

## Support and Maintenance

### Monitoring
- Check GitHub Actions logs for review results
- Monitor API usage in Anthropic Console
- Review generated issues and comments for quality

### Updates
- Update the Claude model version in `.github/workflows/documentation-review.yml`
- Adjust prompts based on review quality
- Modify trigger patterns as project evolves

## Contributing to Claude Integration

### Improving Review Quality
1. **Analyze Review Output**: Look for false positives/negatives
2. **Refine Prompts**: Update prompts in workflow files
3. **Add Custom Focus Areas**: Extend review criteria in config
4. **Test Changes**: Use feature branches to test improvements

### Feedback Loop
- Create issues for Claude integration improvements
- Tag with `claude-integration` label
- Include examples of suboptimal reviews
- Suggest prompt improvements

---

**Note**: This integration requires an Anthropic API key and will incur costs based on usage. Monitor your API usage and adjust triggers as needed to balance cost and review quality.