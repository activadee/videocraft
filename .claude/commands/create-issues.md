# Claude Slash Command: `/create-issues`

## Command Overview
**Purpose**: Analyze a feature idea against current implementation and create GitHub Issues with structured task breakdown.

**Syntax**: `/create-issues <feature-description> [--repo=owner/repo] [--milestone=name] [--epic]`

**Examples**: 
- `/create-issues "Add user authentication with JWT tokens"`
- `/create-issues "Implement real-time notifications" --repo=myorg/backend --milestone="v2.1"`
- `/create-issues "Dark mode theme support" --repo=myorg/frontend --epic`

**Note**: Creates individual GitHub Issues for each task, with proper labels, milestones, and cross-references.

---

## Command Description

This slash command performs comprehensive feature analysis and creates GitHub Issues for implementation. Claude will:

1. **Analyze** the feature request and current system architecture
2. **Break down** the feature into manageable, testable GitHub Issues
3. **Create** structured Issues with labels, milestones, and task lists
4. **Link** related Issues and establish proper dependencies
5. **Generate** an Epic Issue (optional) to track overall feature progress
6. **Link** All Issues to the Epic issues so they have the epic as parent. 
7. **Apply** consistent labeling and project management metadata

---

## System Instructions

### Role Definition
You are a senior technical architect and GitHub project manager specializing in issue creation and project organization. When the `/create-issues` command is invoked, you will:

1. **Switch to GitHub Issues mode**
2. **Conduct thorough feature analysis**
3. **Create properly structured GitHub Issues**
4. **Establish issue relationships and dependencies**
5. **Apply appropriate labels and metadata**

### GitHub Issues Workflow

#### Phase 1: Feature Analysis & Issue Planning
```
STEP 1: Feature Decomposition
- Break down feature into atomic, implementable tasks
- Identify Epic vs Individual Issues structure
- Map dependencies between tasks
- Determine appropriate GitHub labels and milestones

STEP 2: Repository Analysis
- Review existing issue patterns and labels
- Check current milestone structure
- Analyze existing Epic/Feature tracking methods
- Assess team workflow and conventions
```

#### Phase 2: Issue Creation Strategy
```
STEP 3: Issue Structure Planning
- Plan Epic Issue (if --epic flag used)
- Design individual task issues
- Establish issue linking strategy
- Plan label taxonomy and assignments

STEP 4: Content Generation
- Generate issue titles (clear, actionable)
- Create detailed issue descriptions
- Develop acceptance criteria checklists
- Plan issue templates and formatting
```

---

## GitHub Issue Template Structure

### Epic Issue Template (when --epic flag used)
```markdown
# Epic: [Feature Name]

## 🎯 Feature Overview
**Feature Request**: [Original description]
**Business Value**: [Why this feature matters]
**Target Release**: [Milestone/Version]
**Epic Size**: [Estimated total effort]

## 📋 Implementation Tasks
- [ ] #[issue-1] - [Task 1 Title]
- [ ] #[issue-2] - [Task 2 Title]  
- [ ] #[issue-3] - [Task 3 Title]
- [ ] #[issue-4] - [Task 4 Title]

## 🏗️ Architecture Overview
[High-level technical approach and system changes]

## ✅ Epic Acceptance Criteria
- [ ] All subtasks completed and tested
- [ ] Integration testing passed
- [ ] Documentation updated
- [ ] Performance requirements met
- [ ] Security review completed

## 🔗 Related Issues
- Depends on: [List of blocking issues/epics]
- Blocks: [List of dependent issues/epics]

## 📊 Progress Tracking
**Status**: Planning
**Completion**: 0/[total-tasks] tasks completed
**Last Updated**: [Auto-updated timestamp]

---
**Labels**: `epic`, `feature`, `[priority-level]`, `[component-area]`
**Milestone**: [Target milestone]
**Assignee**: [Team lead or architect]
```

### Individual Task Issue Template
```markdown
# [Task Title]: [Clear, Action-Oriented Description]

## 📖 User Story
As a [user type], I want [functionality] so that [business value/benefit].

## ✅ Acceptance Criteria
- [ ] **Given** [context/initial state]  
      **When** [user action/trigger event]  
      **Then** [expected outcome/system behavior]

- [ ] **Given** [alternative context]  
      **When** [different action]  
      **Then** [alternative outcome]

- [ ] **Given** [edge case scenario]  
      **When** [error condition]  
      **Then** [proper error handling]

## 🔧 Technical Specifications
### Implementation Approach
[Technical strategy and approach]

### API Changes
- [ ] [Endpoint 1]: [Description]
- [ ] [Endpoint 2]: [Description]

### Database Changes
- [ ] [Schema change 1]: [Description]
- [ ] [Migration script]: [Description]

### Security Considerations
- [ ] [Security requirement 1]
- [ ] [Security requirement 2]

## 🧪 Testing Requirements
### Unit Tests
- [ ] [Test scenario 1]
- [ ] [Test scenario 2]

### Integration Tests  
- [ ] [Integration test 1]
- [ ] [Integration test 2]

### E2E Tests
- [ ] [User workflow test 1]
- [ ] [User workflow test 2]

## 📚 Definition of Done
- [ ] All acceptance criteria implemented and tested
- [ ] Code review completed and approved
- [ ] Unit tests written and passing (>90% coverage)
- [ ] Integration tests passing
- [ ] API documentation updated
- [ ] Security review completed (if applicable)
- [ ] Performance benchmarks met
- [ ] Error handling and logging implemented
- [ ] Feature flag configuration (if applicable)

## 🔗 Dependencies
**Blocked by**: [List of issues that must be completed first]
**Blocks**: [List of issues that depend on this one]
**Related**: [List of related issues for context]

## 💡 Implementation Notes
**Estimated Effort**: [Story points/hours]
**Priority**: [High/Medium/Low]
**Complexity**: [High/Medium/Low]
**Risk Level**: [High/Medium/Low]

### Technical Decisions
- [Decision 1]: [Rationale]
- [Decision 2]: [Rationale]

### Reference Materials
- [Design document link]
- [API specification link]
- [Research findings link]

---
**Labels**: `task`, `[priority-level]`, `[component-area]`, `[task-type]`
**Milestone**: [Target milestone]
**Epic**: #[epic-issue-number]
**Estimated Time**: [Hours/Story Points]
```

---

## Label Strategy & Organization

### Automatic Label Application
```yaml
# Priority Labels
priority/critical: Issues blocking release or causing system failures
priority/high: Important features or significant bugs
priority/medium: Standard feature work and non-critical bugs  
priority/low: Nice-to-have improvements and minor issues

# Type Labels
type/epic: Large features spanning multiple issues
type/task: Individual implementation tasks
type/bug: Defect fixes
type/enhancement: Improvements to existing features
type/documentation: Documentation updates

# Component Labels (Auto-detected from codebase analysis)
component/api: Backend API changes
component/frontend: UI/UX implementation
component/database: Database schema or query changes
component/auth: Authentication and authorization
component/integration: Third-party service integration
component/testing: Test infrastructure and quality assurance

# Status Labels
status/planning: Issue in planning phase
status/ready: Ready for development
status/in-progress: Currently being worked on
status/review: In code review
status/testing: In QA testing phase
status/blocked: Blocked by dependencies

# Effort Labels
effort/small: 1-2 days of work
effort/medium: 3-5 days of work  
effort/large: 1-2 weeks of work
effort/xlarge: 2+ weeks of work
```

### Milestone Integration
```markdown
# Milestone Strategy
- **Current Sprint**: Issues for immediate sprint
- **Next Sprint**: Planned for upcoming sprint
- **v[X.Y]**: Release-specific milestone
- **Backlog**: Not yet scheduled
- **Research**: Investigation and proof-of-concept work
```

---

## Command Behavior & GitHub Integration

### Repository Detection & Configuration
1. **Auto-detect Repository**: Parse git remote or use current working directory
2. **GitHub API Integration**: Use GitHub CLI or API tokens for issue creation
3. **Permission Validation**: Verify write access to target repository
4. **Template Compliance**: Follow repository's issue templates if they exist

### Issue Creation Process
```
STEP 1: Validate GitHub Access
- Check GitHub CLI authentication
- Verify repository permissions
- Validate milestone and label existence

STEP 2: Create Epic Issue (if --epic flag)
- Generate epic with feature overview
- Create task checklist with placeholders
- Apply appropriate labels and milestone

STEP 3: Create Individual Task Issues
- Generate each task issue with detailed specifications
- Link to epic issue (if applicable)
- Establish dependency relationships
- Apply labels based on content analysis

STEP 4: Update Cross-References
- Update epic issue with created task issue numbers
- Add dependency links between related issues
- Update project boards (if configured)
```

### Advanced GitHub Features
```markdown
# Project Board Integration
- Automatically add issues to project boards
- Set appropriate status columns
- Configure automation rules

# Issue Templates Integration  
- Respect existing repository issue templates
- Merge generated content with template structure
- Maintain repository conventions

# GitHub Actions Integration
- Trigger workflows on issue creation
- Auto-assign based on component labels
- Update related documentation
```

---

## Usage Examples with GitHub Integration

### Example 1: Simple Feature with Auto-Detection
```bash
/create-issues "Add user authentication with JWT tokens"
```
**GitHub Actions**:
1. Creates Epic: "Epic: User Authentication System" 
2. Creates 8 individual task issues:
   - `#123: Implement JWT token generation service`
   - `#124: Create user registration endpoint` 
   - `#125: Add login authentication flow`
   - `#126: Implement password reset functionality`
   - `#127: Add role-based access control`
   - `#128: Create authentication middleware`
   - `#129: Implement logout and token revocation`
   - `#130: Add authentication integration tests`
3. **Labels Applied**: `epic`, `feature`, `priority/high`, `component/auth`
4. **Cross-references**: Epic contains checklist of all task issues

### Example 2: Specific Repository and Milestone
```bash
/create-issues "Implement real-time notifications" --repo=myorg/backend --milestone="v2.1"
```
**GitHub Actions**:
1. Creates issues in `myorg/backend` repository
2. Assigns all issues to "v2.1" milestone  
3. Creates 12 task issues covering WebSocket infrastructure, message queuing, etc.
4. **Labels Applied**: `feature`, `priority/medium`, `component/api`, `component/integration`

### Example 3: Epic Creation for Large Feature
```bash
/create-issues "Complete checkout and payment system" --epic --milestone="Q2-2025"
```
**GitHub Actions**:
1. Creates Epic issue with comprehensive overview
2. Creates 20+ individual task issues covering:
   - Shopping cart functionality
   - Payment provider integration  
   - Order management system
   - Invoice generation
   - Refund processing
3. **Project Board**: Adds epic to "Major Features" project
4. **Dependencies**: Maps task dependencies and blockers

---

## Integration with `/work-on-task`

### Modified `/work-on-task` for GitHub Issues
```bash
# Work on specific GitHub issue
/work-on-task --issue=123
/work-on-task --issue=myorg/repo#123

# Work on issue with local checkout
/work-on-task --issue=123 --branch=feature/auth-jwt-tokens
```

### Enhanced GitHub Workflow
```markdown
# Complete Feature Development Workflow

1. **Feature Planning**:
   /create-issues "Add user authentication" --epic --milestone="v2.0"

2. **Implementation**:
   /work-on-task --issue=123  # Start TDD on first task
   
3. **Automatic Updates**:
   - Issue status automatically updated to "in-progress"
   - Commits reference issue numbers
   - PR creation links back to issues
   
4. **Completion Tracking**:
   - Checkboxes in epic automatically updated
   - Progress tracking across all related issues
   - Milestone progress visualization
```

---

## Error Handling & GitHub Integration

### GitHub Authentication Issues
```
❌ Error: GitHub authentication required.
Please run: gh auth login
Or set GITHUB_TOKEN environment variable.
```

### Repository Access Issues  
```
❌ Error: No write access to repository 'myorg/backend'.
Please check repository permissions or specify a different repository with --repo flag.
```

### Milestone/Label Issues
```
⚠️  Warning: Milestone 'v2.1' not found in repository.
Created milestone 'v2.1' with default due date.

⚠️  Warning: Label 'component/auth' not found.
Created label 'component/auth' with default color.
```

### Rate Limiting
```
⚠️  Warning: GitHub API rate limit approaching.
Created 8 of 12 planned issues. Remaining 4 issues queued for retry in 60 minutes.
Use --batch flag to queue all issues for background creation.
```

---

## Advanced Features & Configuration

### Repository Configuration File
```yaml
# .github/claude-config.yml
issue_creation:
  default_labels:
    - "needs-triage"
    - "created-by-ai"
  
  auto_assign:
    component/auth: "@auth-team"
    component/frontend: "@ui-team"
    component/api: "@backend-team"
  
  project_boards:
    epic: "Major Features"
    task: "Development Backlog"
  
  milestone_strategy: "auto-create"
  
  templates:
    epic: ".github/ISSUE_TEMPLATE/epic.md"
    task: ".github/ISSUE_TEMPLATE/task.md"
```

### Batch Creation Mode
```bash
/create-issues "Large feature" --batch --async
```
**Benefits**:
- Avoids GitHub API rate limits
- Creates issues in background
- Provides progress updates
- Handles failures gracefully

---

## Benefits of GitHub Issues Approach

### ✅ **Better Project Management**
- **Native Integration**: Works with existing GitHub workflow
- **Project Boards**: Automatic kanban board updates
- **Milestone Tracking**: Built-in progress visualization
- **Team Collaboration**: Comments, assignments, reviews

### ✅ **Enhanced Visibility**
- **Searchable History**: All tasks searchable in GitHub
- **Cross-referencing**: Automatic linking between PRs and issues
- **Notifications**: Team members get relevant updates
- **Integration**: Works with existing CI/CD and automation

### ✅ **Improved Workflow**
- **Branch Linking**: Automatic branch-to-issue association
- **PR Integration**: Pull requests automatically close issues
- **Status Tracking**: Real-time progress on features
- **Historical Record**: Complete audit trail of development

### ✅ **Better Tooling**
- **GitHub CLI**: Direct terminal integration
- **API Access**: Programmatic issue management
- **Third-party Tools**: Integration with project management tools
- **Mobile Access**: GitHub mobile app support

---

*This command creates production-ready GitHub Issues optimized for team collaboration and TDD implementation using the `/work-on-task` workflow.*