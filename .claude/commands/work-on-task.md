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
STEP 6: Update GitHub Issue
- Post progress comments with implementation details
- Update issue checkboxes as requirements are completed
- Link commits to issue using conventional commit messages
- Update issue labels (in-progress → review → done)
- Request code review if implementation complete
- Close issue when all acceptance criteria pass
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

### Issue Completion & Review Request
```
✅ **Implementation Complete - Ready for Review**

**All Acceptance Criteria Met**:
- [X] [Criteria 1] - Tests passing
- [X] [Criteria 2] - Tests passing  
- [X] [Criteria 3] - Tests passing

**Code Quality**:
- [X] Test coverage: [X]%
- [X] All existing tests still passing
- [X] Code follows project standards
- [X] Documentation updated

**Review Checklist**:
- [X] Functional requirements met
- [X] Security considerations addressed
- [X] Performance requirements satisfied
- [X] Error handling comprehensive

🔀 **Pull Request**: #[PR-number]  
👥 **Ready for Code Review**
**Progress**: 100% - Complete
```
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
3. **Authentication**: GitHub CLI authenticated or token available
4. **Branch Permissions**: Ability to create and push branches

### Issue Format Requirements
1. **Acceptance Criteria**: Issue must contain testable acceptance criteria
2. **Technical Specs**: Implementation details should be provided
3. **Definition of Done**: Clear completion requirements
4. **Proper Labels**: Appropriate type and component labels

### TDD Enforcement Rules
1. **Always start with failing tests**: No implementation before tests
2. **Test every acceptance criterion**: 1:1 mapping minimum
3. **Incremental development**: Small, focused commits
4. **Continuous validation**: Run tests after each change
5. **GitHub integration**: Commit messages reference issue number

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
6. **Completion**: Issue checkboxes updated, ready for review

### Scenario 2: Cross-Repository Issue
```bash
/work-on-task myorg/backend#456
```
**Expected Workflow**:
1. **Repository Switch**: Validates access to `myorg/backend`
2. **Issue Analysis**: Fetches issue #456 from specified repository
3. **Local Setup**: Ensures local repo is synced with remote
4. **Implementation**: Standard TDD workflow with cross-repo issue updates

### Scenario 3: Custom Branch Strategy
```bash
/work-on-task 789 --branch=feature/auth-jwt-tokens
```
**Expected Workflow**:
1. **Custom Branch**: Creates specified branch name instead of auto-generated
2. **Issue Linking**: Links custom branch to issue #789
3. **Implementation**: Normal TDD workflow with custom branch
4. **PR Creation**: Pull request from custom branch references issue

### Scenario 4: Epic Sub-task Implementation
```bash
/work-on-task 234
# Where #234 is part of Epic #200
```
**Expected Workflow**:
1. **Epic Context**: Recognizes issue is part of larger epic
2. **Epic Updates**: Updates epic progress when sub-task completed
3. **Cross-referencing**: Maintains links between epic and sub-tasks
4. **Milestone Tracking**: Updates milestone progress automatically

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

### GitHub API Rate Limits
```
⚠️  Warning: GitHub API rate limit reached.
- Issue updates will be queued for retry
- Local development can continue
- Comments will be posted when rate limit resets
- Use --offline flag to skip GitHub integration
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
# Auto-generated PR description
## Description
Implements GitHub Issue #123: Add JWT authentication middleware

## Implementation Summary
- ✅ JWT token validation middleware
- ✅ Role-based access control
- ✅ Comprehensive error handling
- ✅ Unit and integration tests

## Testing
- All acceptance criteria tests passing
- Code coverage: 95%
- Security review completed

## Related Issues
Closes #123
Part of Epic #100

## Checklist
- [x] All tests passing
- [x] Code review completed
- [x] Documentation updated
- [x] Security considerations addressed
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

### TDD with GitHub Integration
1. **Test First**: Write failing tests before any implementation
2. **Commit Frequently**: Small commits with clear messages
3. **Issue Updates**: Update checkboxes as requirements are completed
4. **Review Ready**: Mark issue ready for review when all criteria met
5. **Documentation**: Update issue with implementation notes and decisions

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