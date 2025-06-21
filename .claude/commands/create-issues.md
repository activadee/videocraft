# Claude Code Command: `/create-issues`

## Command Overview
**Purpose**: Analyze a feature idea against current implementation and create a single comprehensive GitHub Issue with structured task breakdown.

**Syntax**: `/create-issues <feature-description> [--repo=owner/repo] [--milestone=name] [--type=feature|enhancement|task]`

**Examples**: 
- `/create-issues "Add user authentication with JWT tokens"`
- `/create-issues "Implement real-time notifications" --repo=myorg/backend --milestone="v2.1"`
- `/create-issues "Dark mode theme support" --repo=myorg/frontend --type=enhancement`

**Note**: Creates a single comprehensive GitHub Issue with detailed task breakdown, proper labels, milestone, and implementation plan.

---

## Command Description

This slash command performs comprehensive feature analysis and creates a single structured GitHub Issue for implementation. Claude will:

1. **Analyze** the feature request and current system architecture
1.1 **Ultrathink** about three different approchaes of implementations with pro and cons, and choose the best one according our Architecture. Use Web Search to find the best practice and MCP tools to discuss your solution with different llms. 
   Collaborative tools:
   • mcp__multi-ai-collab__ask_all_ais - Ask all AIs the same question
   • mcp__multi-ai-collab__ai_debate - Have two AIs debate a topic
   • mcp__multi-ai-collab__server_status - Check which AIs are available

   Individual AI tools:
   • mcp__multi-ai-collab__ask_[ai_name]
   • mcp__multi-ai-collab__[ai_name]_code_review
2. **Create** a comprehensive Issue with detailed task breakdown
3. **Structure** the issue with clear sections, checklists, and acceptance criteria
4. **Apply** appropriate labels, milestones, and metadata
5. **Include** implementation roadmap and testing requirements
6. **Establish** clear definition of done and success criteria

---

## System Instructions

### Role Definition
You are a senior technical architect and GitHub project manager specializing in comprehensive issue creation and project organization. When the `/create-issues` command is invoked, you will:

1. **Switch to GitHub Issues mode**
2. **Conduct thorough feature analysis**
3. **Create a single, detailed GitHub Issue**
4. **Structure comprehensive implementation plan**
5. **Apply appropriate labels and metadata**

### Single Issue Workflow

#### Phase 1: Feature Analysis & Planning
```
STEP 1: Documentation Review (REQUIRED FIRST)
- Read docs/ folder for project architecture and guidelines
- Review llms.txt and llms-full.txt for LLM-friendly documentation
- Understand existing patterns, conventions, and standards
- Identify relevant technical documentation and examples

STEP 2: Feature Decomposition
- Analyze feature requirements and complexity
- Break down into logical implementation phases
- Identify all technical components and dependencies
- Map out testing and documentation requirements

STEP 3: Repository Analysis
- Review existing issue patterns and labels
- Check current milestone structure
- Assess team workflow and conventions
- Determine appropriate categorization
```

#### Phase 2: Comprehensive Issue Creation
```
STEP 4: Issue Structure Design
- Plan comprehensive issue sections
- Design detailed task breakdown
- Establish clear acceptance criteria
- Plan implementation roadmap

STEP 5: Content Generation
- Generate clear, actionable issue title
- Create detailed issue description
- Develop comprehensive task checklist
- Include testing and quality requirements
```

---

## GitHub Issue Template Structure

### Comprehensive Feature Issue Template
```markdown
# [Feature Name]: [Clear, Action-Oriented Description]

## 🎯 Feature Overview
**Feature Request**: [Original description]
**Business Value**: [Why this feature matters and impact on users]
**Priority**: [Critical/High/Medium/Low]
**Complexity**: [High/Medium/Low - with brief justification]
**Estimated Effort**: [Story points/hours/days]

## 📖 User Story
As a [user type], I want [functionality] so that [business value/benefit].

**Additional User Scenarios**:
- As a [secondary user type], I need [specific functionality] to [achieve goal]
- As a [admin/power user], I want [advanced functionality] for [administrative purpose]

## ✅ Acceptance Criteria
### Core Functionality
- [ ] **Given** [context/initial state]  
      **When** [user action/trigger event]  
      **Then** [expected outcome/system behavior]

- [ ] **Given** [alternative context]  
      **When** [different user action]  
      **Then** [alternative expected outcome]

### Edge Cases & Error Handling
- [ ] **Given** [error condition/edge case]  
      **When** [problematic scenario occurs]  
      **Then** [appropriate error handling/user feedback]

- [ ] **Given** [boundary condition]  
      **When** [limit is reached/exceeded]  
      **Then** [graceful degradation/proper messaging]

### Performance & Security
- [ ] Feature performs within [X seconds/ms] under normal load
- [ ] Security requirements met (authentication, authorization, data protection)
- [ ] Accessibility standards followed (WCAG 2.1 AA compliance)
- [ ] Mobile responsiveness maintained across devices

## 🔧 Implementation Plan

### Phase 1: Foundation & Setup
- [ ] **Database Changes**
  - [ ] [Schema modification 1]: [Description and SQL migration]
  - [ ] [Schema modification 2]: [Description and indexing strategy]
  - [ ] [Data migration script]: [Handle existing data transformation]

- [ ] **API Design & Backend**
  - [ ] [Endpoint 1]: `[METHOD] /api/path` - [Purpose and specification]
  - [ ] [Endpoint 2]: `[METHOD] /api/path` - [Purpose and specification]
  - [ ] [Service layer]: [Business logic implementation]
  - [ ] [Authentication/Authorization]: [Security implementation]

### Phase 2: Frontend Implementation
- [ ] **UI Components**
  - [ ] [Component 1]: [Description and functionality]
  - [ ] [Component 2]: [Description and user interactions]
  - [ ] [Page/View updates]: [Navigation and layout changes]

- [ ] **State Management**
  - [ ] [State structure]: [Data flow and management approach]
  - [ ] [API integration]: [Frontend-backend communication]
  - [ ] [Error handling]: [User feedback and retry mechanisms]

### Phase 3: Integration & Polish
- [ ] **Feature Integration**
  - [ ] [System integration 1]: [How feature connects with existing functionality]
  - [ ] [Third-party integration]: [External service setup and configuration]
  - [ ] [Notification system]: [User alerts and communication]

- [ ] **User Experience**
  - [ ] [Loading states]: [Progressive loading and skeleton screens]
  - [ ] [Empty states]: [First-time user experience]
  - [ ] [Onboarding]: [Feature introduction and guidance]

## 🧪 Testing Strategy

### Unit Tests
- [ ] **Backend Tests**
  - [ ] [Service class tests]: [Business logic validation]
  - [ ] [Repository tests]: [Data access layer testing]
  - [ ] [Utility function tests]: [Helper method validation]

- [ ] **Frontend Tests**
  - [ ] [Component tests]: [UI component behavior testing]
  - [ ] [Hook tests]: [Custom hook functionality]
  - [ ] [Utility tests]: [Frontend helper functions]

### Integration Tests
- [ ] **API Integration**
  - [ ] [Endpoint integration tests]: [Full request/response cycle]
  - [ ] [Database integration]: [Data persistence and retrieval]
  - [ ] [Authentication flow]: [Security integration testing]

- [ ] **Frontend Integration**
  - [ ] [User workflow tests]: [Complete user journey testing]
  - [ ] [API communication]: [Frontend-backend integration]
  - [ ] [State management]: [Data flow integration testing]

### End-to-End Tests
- [ ] **Critical User Paths**
  - [ ] [Primary user workflow]: [Main feature usage scenario]
  - [ ] [Alternative workflows]: [Secondary usage patterns]
  - [ ] [Error recovery]: [User error handling and recovery]

- [ ] **Cross-Browser Testing**
  - [ ] [Chrome/Firefox/Safari]: [Core browser compatibility]
  - [ ] [Mobile browsers]: [Mobile device functionality]
  - [ ] [Accessibility testing]: [Screen reader and keyboard navigation]

## 🔒 Security Considerations
- [ ] **Authentication & Authorization**
  - [ ] [Permission checks]: [Role-based access control implementation]
  - [ ] [Data access validation]: [User can only access authorized data]
  - [ ] [API security]: [Input validation and sanitization]

- [ ] **Data Protection**
  - [ ] [Sensitive data handling]: [Encryption and secure storage]
  - [ ] [Privacy compliance]: [GDPR/CCPA considerations]
  - [ ] [Audit logging]: [Security event tracking]

## 📊 Performance Requirements
- [ ] **Response Times**
  - [ ] [API endpoints]: [< X ms response time under normal load]
  - [ ] [Page load times]: [< X seconds for feature pages]
  - [ ] [Database queries]: [Optimized query performance]

- [ ] **Scalability**
  - [ ] [Concurrent users]: [Handle X simultaneous users]
  - [ ] [Data volume]: [Performance with large datasets]
  - [ ] [Resource usage]: [Memory and CPU optimization]

## 📚 Documentation Requirements
- [ ] **Technical Documentation**
  - [ ] [API documentation]: [OpenAPI/Swagger specification updates]
  - [ ] [Database schema]: [ERD and migration documentation]
  - [ ] [Architecture decisions]: [ADR for major technical choices]

- [ ] **User Documentation**
  - [ ] [Feature guide]: [User-facing documentation]
  - [ ] [Help content]: [In-app help and tooltips]
  - [ ] [FAQ updates]: [Common questions and answers]

## 🚀 Deployment & Release
- [ ] **Deployment Preparation**
  - [ ] [Environment configuration]: [Production environment setup]
  - [ ] [Feature flags]: [Gradual rollout configuration]
  - [ ] [Monitoring setup]: [Logging and alerting configuration]

- [ ] **Release Planning**
  - [ ] [Release notes]: [User-facing change documentation]
  - [ ] [Rollback plan]: [Reversion strategy if issues arise]
  - [ ] [Success metrics]: [KPIs to measure feature success]

## ✅ Definition of Done
### Code Quality
- [ ] All acceptance criteria implemented and tested
- [ ] Code review completed and approved by 2+ team members
- [ ] Unit tests written and passing (>90% coverage for new code)
- [ ] Integration tests passing
- [ ] E2E tests covering critical paths

### Documentation & Communication
- [ ] API documentation updated (if applicable)
- [ ] User documentation created/updated
- [ ] Architecture decisions documented
- [ ] Release notes prepared

### Quality Assurance
- [ ] Manual testing completed across browsers/devices
- [ ] Performance benchmarks met
- [ ] Security review completed (if applicable)
- [ ] Accessibility audit passed

### Production Readiness
- [ ] Feature flag configuration completed
- [ ] Monitoring and alerting configured
- [ ] Database migrations tested in staging
- [ ] Rollback procedures documented and tested

## 🔗 Dependencies & Blockers
**Blocked by**: 
- [ ] [Prerequisite 1]: [Description of dependency]
- [ ] [Prerequisite 2]: [External team/resource dependency]

**Blocks**: 
- [ ] [Future work 1]: [Work that depends on this feature]
- [ ] [Future work 2]: [Related features that need this foundation]

**Related Issues**: 
- [Issue #X]: [Related feature or bug that impacts this work]
- [Issue #Y]: [Complementary work or shared components]

## 💡 Technical Decisions & Notes
### Architecture Decisions
- **[Decision 1]**: [Technology/approach choice] - [Rationale and trade-offs]
- **[Decision 2]**: [Design pattern/framework choice] - [Benefits and considerations]

### Implementation Notes
- **[Note 1]**: [Important implementation detail or constraint]
- **[Note 2]**: [Performance consideration or optimization opportunity]

### Risk Assessment
- **High Risk**: [Major technical or business risk] - [Mitigation strategy]
- **Medium Risk**: [Moderate risk factor] - [Monitoring and contingency plan]

## 📋 Subtask Checklist
### Setup & Planning
- [ ] Requirements review with stakeholders
- [ ] Technical design document created
- [ ] Database migration scripts written
- [ ] Development environment setup

### Backend Development
- [ ] Data models implemented
- [ ] Business logic services created
- [ ] API endpoints developed
- [ ] Authentication/authorization integrated

### Frontend Development
- [ ] UI components built
- [ ] State management implemented
- [ ] API integration completed
- [ ] User interactions polished

### Testing & Quality
- [ ] Unit tests implemented
- [ ] Integration tests created
- [ ] E2E tests developed
- [ ] Manual testing completed

### Documentation & Release
- [ ] Code documentation updated
- [ ] User guides created
- [ ] Deployment scripts prepared
- [ ] Feature deployed to production

## 📈 Success Metrics
- **User Adoption**: [Specific metric] within [timeframe]
- **Performance**: [Specific performance target]
- **Business Impact**: [Measurable business outcome]

---
**Labels**: `feature`, `[priority-level]`, `[component-area]`, `[effort-size]`
**Milestone**: [Target milestone]
**Assignee**: [Team member or team tag]
**Epic**: [Parent epic if applicable]
**Estimated Time**: [Story points/hours/days]
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
type/feature: New functionality or major enhancements
type/enhancement: Improvements to existing features
type/task: Implementation work or technical tasks
type/bug: Defect fixes (if combined with feature work)

# Component Labels (Auto-detected from codebase analysis)
component/api: Backend API changes
component/frontend: UI/UX implementation
component/database: Database schema or query changes
component/auth: Authentication and authorization
component/integration: Third-party service integration
component/testing: Test infrastructure and quality assurance

# Effort Labels
effort/small: 1-3 days of work
effort/medium: 4-7 days of work  
effort/large: 1-2 weeks of work
effort/xlarge: 2+ weeks of work

# Status Labels
status/planning: Issue in planning phase
status/ready: Ready for development
status/in-progress: Currently being worked on
status/review: In code review
status/blocked: Blocked by dependencies
```

---

## Command Behavior & GitHub Integration

### Issue Creation Process
```
STEP 1: Validate GitHub Access
- Check GitHub CLI authentication
- Verify repository permissions
- Validate milestone and label existence

STEP 2: Analyze Feature Complexity
- Determine appropriate issue template
- Assess effort and complexity levels
- Identify required components and integrations

STEP 3: Create Comprehensive Issue
- Generate detailed issue with all sections
- Apply appropriate labels based on content analysis
- Set milestone and assignee (if specified)
- Link to related issues or epics (if applicable)

STEP 4: Provide Summary
- Display created issue URL and number
- Show applied labels and metadata
- Provide next steps for implementation
```

---

## Usage Examples

### Example 1: Simple Feature
```bash
/create-issues "Add user authentication with JWT tokens"
```
**GitHub Actions**:
1. Creates comprehensive issue: "Feature: User Authentication with JWT Tokens"
2. **Labels Applied**: `feature`, `priority/high`, `component/auth`, `effort/large`
3. **Sections Include**: 15+ implementation tasks, testing strategy, security considerations

### Example 2: Specific Repository and Milestone
```bash
/create-issues "Implement real-time notifications" --repo=myorg/backend --milestone="v2.1"
```
**GitHub Actions**:
1. Creates issue in `myorg/backend` repository
2. Assigns to "v2.1" milestone  
3. **Labels Applied**: `feature`, `priority/medium`, `component/api`, `component/integration`
4. **Comprehensive breakdown**: WebSocket setup, message queuing, notification templates, etc.

### Example 3: Enhancement Type
```bash
/create-issues "Dark mode theme support" --type=enhancement --milestone="Q2-2025"
```
**GitHub Actions**:
1. Creates enhancement issue with UI/UX focus
2. **Labels Applied**: `enhancement`, `priority/medium`, `component/frontend`, `effort/medium`
3. **Detailed tasks**: Theme switching, color system, user preferences, etc.

---

## Integration with `/work-on-task`

### Enhanced GitHub Workflow
```markdown
# Complete Feature Development Workflow

1. **Feature Planning**:
   /create-issues "Add user authentication" --milestone="v2.0"

2. **Implementation**:
   /work-on-task --issue=123  # Start TDD on the comprehensive issue
   
3. **Progress Tracking**:
   - Check off completed subtasks in issue description
   - Update issue status with progress comments
   - Link commits and PRs to the issue
   
4. **Completion**:
   - All checklist items completed
   - PR merged with "Closes #123"
   - Issue automatically closed
```
