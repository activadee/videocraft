# Claude Slash Command: `/work-on-task`

## Command Overview
**Purpose**: Start working on a specific GitHub Issue using Test-Driven Development (TDD) methodology.

**Syntax**: `/work-on-task <issue-reference> [--branch=branch-name] [--repo=owner/repo]`

**Examples**: 
- `/work-on-task 123`
- `/work-on-task myorg/backend#456`
- `/work-on-task 789 --branch=feature/auth-jwt-tokens`
- `/work-on-task #234 --repo=myorg/frontend`

**Note**: Works with GitHub Issues created by `/create-issues` or any existing GitHub Issue with proper acceptance criteria.

---

## Command Description

This slash command instructs Claude to begin implementing a specific GitHub Issue using Test-Driven Development (TDD) approach. Claude will:

1. **Fetch** the GitHub Issue details and extract requirements
2. **Analyze** acceptance criteria and technical specifications
3. **Create** git branch and update Issue status (optional)
4. **Generate tests** based on acceptance criteria (Red phase)
5. **Implement** minimal code to pass tests (Green phase)
6. **Refactor** code while maintaining test coverage (Refactor phase)
7. **Update** GitHub Issue with progress and completion status

---

## System Instructions

### Role Definition
You are a senior software engineer specializing in Test-Driven Development (TDD). When the `/work-on-task` command is invoked, you will:

1. **Immediately switch to TDD mode**
2. **Focus exclusively on the specified task**
3. **Follow strict TDD methodology**
4. **Ensure all acceptance criteria become passing tests**
5. **Provide clear, executable code**

### TDD Workflow Implementation

#### Phase 1: GitHub Issue Analysis & Setup
```
STEP 1: Fetch GitHub Issue
- Retrieve issue details via GitHub API/CLI
- Parse issue title, description, and acceptance criteria
- Extract technical specifications and requirements
- Identify linked issues and dependencies
- Check current issue status and assignee

STEP 2: Repository Setup
- Verify local repository state and remote connection
- Create feature branch (if --branch specified)
- Update issue status to "in-progress" (optional)
- Add assignee and labels as needed
- Post comment indicating work has started
```

#### Phase 2: Acceptance Criteria Mapping
```
STEP 3: Parse Issue Requirements
- Extract "Acceptance Criteria" section from issue
- Convert checkboxes into testable scenarios
- Identify "Definition of Done" requirements
- Map technical specifications to implementation tasks
- Plan test data and mock requirements
```

#### Phase 2: Red Phase (Failing Tests)
```
STEP 3: Write Failing Tests
- Create test files with descriptive names
- Implement tests for each acceptance criterion
- Include unit tests, integration tests as needed
- Ensure all tests fail initially (Red phase)
- Add test documentation and comments
```

#### Phase 3: Green Phase (Minimal Implementation)
```
STEP 4: Implement Minimal Code
- Write the simplest code to make tests pass
- Focus on functionality, not optimization
- Implement only what's needed for green tests
- Avoid over-engineering or premature optimization
```

#### Phase 4: Refactor Phase (Code Improvement)
```
STEP 5: Refactor & Optimize
- Improve code structure and readability
- Remove duplication and technical debt
- Optimize performance where necessary
- Maintain 100% test coverage
- Update documentation
```

#### Phase 5: GitHub Integration & Completion
```
STEP 6: Final Validation & Git Integration
- Run complete test suite and ensure all tests pass
- Verify all acceptance criteria checkboxes completed
- Validate code quality and standards compliance
- Stage all changes for commit

STEP 7: Commit & Push Changes
- Generate comprehensive commit message with issue reference
- Commit all changes with conventional commit format
- Push feature branch to GitHub remote
- Update issue status to "ready-for-review"

STEP 8: Pull Request Creation
- Create PR with auto-generated description
- Link PR to original issue (auto-closes on merge)
- Apply appropriate labels and request reviewers
- Update issue with PR reference and completion status
- Post final completion comment to issue
```

---

## Output Format

### 1. GitHub Issue Summary
```markdown
## Working on GitHub Issue #[Number]: [Issue Title]

**Repository**: [owner/repo]
**Assignee**: [Current assignee]  
**Milestone**: [Target milestone]
**Labels**: [Applied labels]
**Epic**: [Related epic if applicable]

### Issue Description
[Original issue description]

### Acceptance Criteria Analysis
- ✅ [Checkbox 1]: [Description and test approach]
- ✅ [Checkbox 2]: [Description and test approach]  
- ✅ [Checkbox 3]: [Description and test approach]

### Technical Specifications Identified
- **API Changes**: [Endpoints to modify/create]
- **Database Changes**: [Schema modifications needed]
- **Testing Requirements**: [Unit, integration, E2E tests]
- **Security Considerations**: [Auth, validation, etc.]

### Implementation Strategy
[Brief explanation of TDD approach for this specific issue]

### Git Branch Strategy
**Branch**: `feature/issue-[number]-[brief-description]`
**Base Branch**: [main/develop/etc.]
```

### 2. TDD Implementation Phases

#### Red Phase: Failing Tests
```markdown
### 🔴 RED PHASE: Writing Failing Tests

**Test File**: `tests/[feature-name].test.js`

```javascript
// Test implementation
describe('[Feature Name]', () => {
  it('should [specific behavior]', () => {
    // Test code that fails initially
  });
});
```

**Test Results**: ❌ [Number] tests failing (Expected)
```

#### Green Phase: Minimal Implementation
```markdown
### 🟢 GREEN PHASE: Minimal Implementation

**Implementation File**: `src/[feature-name].js`

```javascript
// Minimal code to pass tests
class FeatureName {
  // Implementation
}
```

**Test Results**: ✅ [Number] tests passing
```

#### Refactor Phase: Code Improvement
```markdown
### 🔵 REFACTOR PHASE: Code Optimization

**Refactored Code**: 
- Improved [specific aspect]
- Extracted [common functionality]
- Optimized [performance aspect]

**Final Test Results**: ✅ All tests passing
**Code Coverage**: [Percentage]%
```

### 3. GitHub Issue Updates & Progress Tracking
```markdown
## 📊 GitHub Issue Progress Updates

### Initial Status Update
**Comment Posted to Issue #[number]**:
```
🚀 **Starting TDD Implementation**

**Branch**: `feature/issue-[number]-[description]`
**Approach**: Test-Driven Development
**Acceptance Criteria**: [X] total identified

**Implementation Plan**:
1. ✅ Red Phase: Write failing tests for all acceptance criteria
2. 🔄 Green Phase: Implement minimal code to pass tests  
3. ⏳ Refactor Phase: Optimize and clean up code
4. ✅ Validation: Verify all requirements met

**Progress**: 0% - Starting implementation
```

### Progress Updates During Implementation
**Red Phase Complete**:
```
🔴 **Red Phase Complete** - All tests written and failing as expected

**Tests Created**:
- [X] Unit tests for [component/feature]
- [X] Integration tests for [API endpoints]
- [X] Edge case tests for [error scenarios]

**Next**: Green Phase - Implementing minimal code to pass tests
**Progress**: 25% - Tests established
```

**Green Phase Complete**:
```
🟢 **Green Phase Complete** - All tests now passing

**Implementation Summary**:
- [X] [Feature 1] implemented and tested
- [X] [Feature 2] implemented and tested
- [X] Error handling implemented

**Code Coverage**: [X]%
**Next**: Refactor Phase - Code optimization and cleanup  
**Progress**: 75% - Core functionality complete
```

### 4. Final Commit & Pull Request Creation
```markdown
## 🚀 Implementation Complete - Creating PR

### Final Commit Message
```bash
git add .
git commit -m "feat: implement JWT authentication middleware (#123)

Complete implementation of JWT token validation and role-based access control

**Features:**
- JWT token validation middleware
- Role-based access control (RBAC)  
- Comprehensive error handling
- Password reset functionality
- Session management

**Testing:**
- Unit tests: 45 tests passing
- Integration tests: 12 tests passing
- E2E tests: 8 tests passing
- Code coverage: 94%

**Security:**
- Input validation implemented
- SQL injection prevention
- XSS protection
- Rate limiting added

**Performance:**
- Token validation: <5ms average
- Database queries optimized
- Caching implemented for user roles

Closes #123"
```

### Pull Request Auto-Generation
**PR Title**: `feat: implement JWT authentication middleware (#123)`

**PR Description**:
```markdown
## 🎯 Description
Implements comprehensive JWT authentication system as specified in Issue #123.

## 📋 Implementation Summary
### ✅ Completed Features
- [x] JWT token generation and validation
- [x] User registration and login endpoints
- [x] Password reset functionality with email verification
- [x] Role-based access control (RBAC)
- [x] Authentication middleware for protected routes
- [x] Logout and token revocation
- [x] Session management and security

### 🧪 Testing Coverage
- **Unit Tests**: 45 tests (100% of new code)
- **Integration Tests**: 12 API endpoint tests
- **E2E Tests**: 8 user workflow tests
- **Code Coverage**: 94% overall
- **Security Tests**: Authentication flow validation

### 🔒 Security Implementation
- [x] Input validation and sanitization
- [x] SQL injection prevention
- [x] XSS protection headers
- [x] Rate limiting on auth endpoints
- [x] Secure password hashing (bcrypt)
- [x] JWT token expiration and refresh
- [x] HTTPS-only session cookies

### 📊 Performance Metrics
- Token validation: <5ms average response time
- Login endpoint: <200ms average response time
- Database queries optimized with proper indexing
- Redis caching for user roles and permissions

## 🔗 Related Issues
Closes #123
Part of Epic #100: Complete Authentication System

## 🔍 Review Checklist
- [x] All acceptance criteria implemented and tested
- [x] Code follows project standards and conventions
- [x] Security review completed - no vulnerabilities found
- [x] Performance requirements met (<200ms response times)
- [x] Error handling comprehensive and user-friendly
- [x] Documentation updated (API docs, README)
- [x] Database migrations tested and validated
- [x] Backward compatibility maintained

## 🧪 Testing Instructions
### Prerequisites
```bash
npm install
cp .env.example .env
# Update .env with test database credentials
npm run db:migrate
```

### Run Test Suite
```bash
# All tests
npm test

# Specific test categories
npm run test:unit
npm run test:integration
npm run test:e2e
npm run test:security
```

### Manual Testing
1. **Registration**: POST `/api/auth/register`
2. **Login**: POST `/api/auth/login`
3. **Protected Route**: GET `/api/profile` (with JWT token)
4. **Password Reset**: POST `/api/auth/reset-password`
5. **Logout**: POST `/api/auth/logout`

## 📚 Documentation Updates
- [x] API documentation updated with new endpoints
- [x] Authentication flow diagram added
- [x] Security considerations documented
- [x] Environment variable documentation updated
- [x] Deployment guide updated with new requirements

## 🚀 Deployment Notes
### Environment Variables Required
```env
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION=24h
BCRYPT_ROUNDS=12
REDIS_URL=redis://localhost:6379
EMAIL_SERVICE_API_KEY=your-email-service-key
```

### Database Migrations
```bash
# Run migrations before deployment
npm run db:migrate

# Rollback if needed
npm run db:rollback
```

### Performance Impact
- New endpoints add ~2ms to application startup
- Redis dependency required for session caching
- Database requires 3 new tables (users, roles, sessions)

---
**Ready for Review** 👥
**Estimated Review Time**: 2-3 hours
**Merge Target**: `main` branch
```

### GitHub CLI Commands Executed
```bash
# Push feature branch
git push origin feature/issue-123-jwt-auth-middleware

# Create pull request
gh pr create \
  --title "feat: implement JWT authentication middleware (#123)" \
  --body-file pr-description.md \
  --label "enhancement,security,backend" \
  --milestone "v2.0" \
  --assignee "@auth-team" \
  --reviewer "@senior-dev,@security-lead"

# Update original issue
gh issue comment 123 --body "✅ **Implementation Complete**

🔀 **Pull Request Created**: #$(gh pr view --json number -q .number)
👥 **Ready for Code Review**
📊 **All Acceptance Criteria Met**

**Next Steps**:
1. Code review by assigned reviewers
2. Security review (if required)
3. QA testing in staging environment
4. Merge to main and deploy

**Implementation Summary**:
- All acceptance criteria implemented and tested
- Comprehensive test suite (94% coverage)
- Security best practices followed
- Performance requirements exceeded
- Documentation updated

Thank you for the clear requirements! 🎉"
```

### Final Issue Status Update
- **Status**: Changed from "in-progress" to "ready-for-review"
- **Labels**: Added "ready-for-review", "awaiting-merge"
- **Assignee**: Updated to reviewers
- **All Checkboxes**: Marked as completed ✅
- **PR Link**: Added to issue description
```

---

## Command Behavior Rules

### GitHub Issue Integration
1. **Issue Validation**: Verify issue exists and is accessible
2. **Permission Check**: Ensure write access to repository
3. **Status Management**: Update issue status appropriately
4. **Branch Creation**: Create feature branch following naming conventions
5. **Progress Tracking**: Post regular updates to issue comments

### Repository Requirements
1. **Git Repository**: Must be in a valid git repository
2. **GitHub Remote**: Repository must have GitHub remote configured
3. **Authentication**: GitHub CLI authenticated with PR creation permissions
4. **Branch Permissions**: Ability to create branches and open PRs
5. **Clean Working State**: All changes properly staged for final commit

### Issue Format Requirements
1. **Acceptance Criteria**: Issue must contain testable acceptance criteria
2. **Technical Specs**: Implementation details should be provided
3. **Definition of Done**: Clear completion requirements
4. **Proper Labels**: Appropriate type and component labels

### TDD Enforcement Rules
1. **Always start with failing tests**: No implementation before tests
2. **Test every acceptance criterion**: 1:1 mapping minimum
3. **Incremental development**: Small, focused commits during development
4. **Continuous validation**: Run tests after each change
5. **Final commit**: Single comprehensive commit with all changes
6. **Automatic PR**: Create pull request with detailed description
7. **Issue closure**: Link PR to issue for automatic closure on merge

---

## Example Usage Scenarios

### Scenario 1: Simple Issue Implementation
```bash
/work-on-task 123
```
**Expected Workflow**:
1. **Fetch Issue**: Gets GitHub Issue #123 from current repository
2. **Branch Creation**: `feature/issue-123-jwt-auth-middleware`
3. **Status Update**: Issue labeled "in-progress", comment posted
4. **TDD Implementation**: Red → Green → Refactor cycle
5. **Progress Updates**: Regular comments on issue with progress
6. **Final Commit**: Comprehensive commit with all changes
7. **PR Creation**: Auto-generated PR with detailed description
8. **Issue Update**: Final completion comment with PR link

### Scenario 2: Cross-Repository Issue
```bash
/work-on-task myorg/backend#456
```
**Expected Workflow**:
1. **Repository Switch**: Validates access to `myorg/backend`
2. **Issue Analysis**: Fetches issue #456 from specified repository
3. **Local Setup**: Ensures local repo is synced with remote
4. **Implementation**: Standard TDD workflow with cross-repo issue updates
5. **PR Creation**: Pull request created in correct repository
6. **Cross-Reference**: Original issue updated with PR link

### Scenario 3: Custom Branch with Auto-PR
```bash
/work-on-task 789 --branch=feature/auth-jwt-tokens
```
**Expected Workflow**:
1. **Custom Branch**: Creates specified branch name instead of auto-generated
2. **Issue Linking**: Links custom branch to issue #789
3. **Implementation**: Normal TDD workflow with custom branch
4. **Auto-Commit**: Final commit with comprehensive message
5. **PR Creation**: Pull request from custom branch with auto-generated description
6. **Reviewer Assignment**: Automatic reviewer assignment based on issue labels

### Scenario 4: Epic Sub-task with Complete Workflow
```bash
/work-on-task 234
# Where #234 is part of Epic #200
```
**Expected Workflow**:
1. **Epic Context**: Recognizes issue is part of larger epic
2. **Implementation**: Standard TDD implementation
3. **Final Commit**: Commit message references both issue and epic
4. **PR Creation**: PR description includes epic context
5. **Epic Updates**: Updates epic progress when PR is created
6. **Milestone Tracking**: Updates milestone progress automatically
7. **Team Notification**: Notifies epic stakeholders of completion

---

## Error Handling

### GitHub Issue Not Found
```
❌ Error: Issue #123 not found in current repository.
Available options:
- Check issue number: /work-on-task 124
- Specify repository: /work-on-task myorg/repo#123
- List open issues: gh issue list
```

### Repository Access Issues
```
❌ Error: No access to repository 'myorg/backend'.
Please check:
- Repository exists and is accessible
- GitHub authentication: gh auth login
- Repository permissions (read/write required)
```

### Issue Format Problems
```
❌ Error: Issue #123 missing acceptance criteria.
GitHub Issues must include:
- Clear acceptance criteria (checkboxes)
- Technical specifications
- Definition of done

Please update issue format or use /create-issues to generate properly formatted issues.
```

### Branch Creation Conflicts
```
❌ Error: Branch 'feature/issue-123-auth' already exists.
Options:
1. Switch to existing branch: git checkout feature/issue-123-auth
2. Use different branch: /work-on-task 123 --branch=feature/auth-v2
3. Delete existing branch: git branch -D feature/issue-123-auth
```

### PR Creation Failures
```
❌ Error: Failed to create pull request.
Possible causes:
- No changes to commit (all acceptance criteria already implemented)
- Branch protection rules prevent PR creation
- Missing required reviewers or approvals
- GitHub API rate limit exceeded

Solutions:
- Verify changes exist: git status
- Check branch protection: gh repo view --web
- Retry with --force flag: /work-on-task 123 --force-pr
```

### Commit Message Generation Issues
```
⚠️  Warning: Unable to generate comprehensive commit message.
Using fallback commit message format.

Fallback commit:
```
feat: implement issue #123

- Complete TDD implementation
- All acceptance criteria met
- Tests passing

Closes #123
```

Recommendation: Review and amend commit message before push.
```

---

## Integration Notes

### GitHub CLI Requirements
- **GitHub CLI**: `gh` command must be installed and authenticated
- **Repository Access**: Read/write permissions to target repository
- **API Tokens**: Personal access token with repo and issues scope

### Git Repository Setup
- **Valid Git Repo**: Must be run from within a git repository
- **GitHub Remote**: Repository must have GitHub remote configured
- **Clean Working Directory**: Uncommitted changes should be stashed

### Commit Message Integration
```bash
# Automatic commit message format
git commit -m "feat: implement JWT auth middleware

- Add token validation middleware
- Implement user role checking
- Add comprehensive error handling

Closes #123"
```

### Pull Request Integration
```markdown
# Auto-generated PR description template
## Description
Implements GitHub Issue #[number]: [issue title]

## Implementation Summary
- ✅ [Feature 1]: [Description]
- ✅ [Feature 2]: [Description]
- ✅ [Feature 3]: [Description]

## Testing
- All acceptance criteria tests passing
- Code coverage: [percentage]%
- [Security/Performance] review completed

## Related Issues
Closes #[issue-number]
Part of Epic #[epic-number] (if applicable)

## Checklist
- [x] All tests passing
- [x] Code review completed
- [x] Documentation updated
- [x] Security considerations addressed
```

### Advanced PR Features
```bash
# Automatic reviewer assignment based on CODEOWNERS
gh pr create --reviewer "@team/backend,@security-lead"

# Auto-assign labels based on issue labels
gh pr create --label "enhancement,security,backend"

# Link to milestone and project boards
gh pr create --milestone "v2.0" --project "Backend Development"

# Draft PR for work-in-progress (optional flag)
/work-on-task 123 --draft-pr
```

### Framework Support
- **JavaScript**: Jest, Mocha, Cypress
- **Python**: pytest, unittest, nose2
- **Java**: JUnit, TestNG, Mockito
- **C#**: NUnit, xUnit, MSTest
- **Go**: testing package, Ginkgo
- **Rust**: built-in test framework

### Continuous Integration
- Tests should be executable in CI/CD pipeline
- Coverage reports should be generated automatically
- Failed tests should block deployment

---

## Best Practices

### GitHub Issue Workflow
1. **Issue Assignment**: Assign yourself when starting work
2. **Branch Naming**: Use consistent `feature/issue-[number]-[description]` format
3. **Regular Updates**: Post progress comments every major milestone
4. **Atomic Commits**: Each commit should reference the issue number
5. **Quality Gates**: Ensure all acceptance criteria pass before completion

### Complete Workflow Integration
1. **Implementation Phase**: TDD cycle with incremental commits
2. **Final Validation**: Comprehensive test suite and acceptance criteria check
3. **Code Consolidation**: Stage all changes for single comprehensive commit
4. **Commit Creation**: Generate detailed commit message with conventional format
5. **Branch Push**: Push feature branch to GitHub remote
6. **PR Generation**: Create pull request with auto-generated description
7. **Issue Completion**: Update original issue with PR link and completion status
8. **Team Notification**: Notify relevant team members and reviewers

### Automated Quality Gates
1. **Test Coverage**: Ensure minimum coverage threshold met
2. **Security Scan**: Run security linting and vulnerability checks  
3. **Code Standards**: Validate code formatting and conventions
4. **Documentation**: Verify API docs and README updates completed
5. **Performance**: Run performance benchmarks if applicable

### Post-Implementation Automation
```bash
# Automatic actions after PR creation
- Issue status updated to "ready-for-review"
- Epic progress updated (if sub-task)
- Milestone progress calculated
- Team notifications sent
- CI/CD pipeline triggered
- Code quality reports generated
```

### Code Quality & Collaboration
1. **Clear Communication**: Detailed progress comments for team visibility
2. **Security First**: Address security considerations in acceptance criteria
3. **Performance Aware**: Include performance requirements in test strategy
4. **Maintainable Code**: Write self-documenting code with proper naming
5. **Team Standards**: Follow existing repository conventions and patterns

### Issue Completion Protocol
1. **All Tests Green**: Comprehensive test suite must pass
2. **Acceptance Criteria**: All checkboxes completed and verified
3. **Code Review**: Request review from appropriate team members
4. **Documentation**: Update relevant documentation and README files
5. **Epic Updates**: Update parent epic progress if applicable

---

*This slash command integrates TDD methodology with GitHub Issues workflow for seamless project management and code quality.*