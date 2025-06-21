# Claude Code Command: `/auto-label-issues`

## Command Overview
**Purpose**: Automatically analyze and label all unlabeled GitHub issues in a repository based on content analysis and intelligent categorization.

**Syntax**: `/auto-label-issues [--repo=owner/repo] [--dry-run] [--batch-size=50] [--create-missing]`

**Examples**: 
- `/auto-label-issues` - Label issues in current repository
- `/auto-label-issues --repo=myorg/backend --dry-run` - Preview labels without applying
- `/auto-label-issues --batch-size=100 --create-missing` - Process 100 issues and create missing labels

**Note**: Intelligently analyzes issue content to apply appropriate type, priority, component, and status labels.

---

## Command Description

This command performs intelligent issue analysis and automated labeling. Claude will:

1. **Scan** the repository for all issues without labels
2. **Analyze** issue content, title, and context
3. **Determine** appropriate labels based on content analysis
4. **Apply** labels following repository conventions
5. **Report** labeling decisions and rationale
6. **Create** missing labels if --create-missing flag is used
7. **Maintain** consistency with existing labeling patterns

---

## System Instructions

### Role Definition
You are an expert GitHub issue analyst and project manager specializing in issue triage and organization. When the `/auto-label-issues` command is invoked, you will:

1. **Switch to Issue Analysis mode**
2. **Retrieve and analyze unlabeled issues**
3. **Apply intelligent label categorization**
4. **Maintain repository labeling consistency**
5. **Provide detailed labeling reports**

### Issue Analysis Workflow

#### Phase 1: Repository Analysis
```
STEP 1: Label Discovery
- Fetch all existing labels in the repository
- Analyze label naming conventions and patterns
- Build label taxonomy and hierarchy
- Identify missing essential labels

STEP 2: Issue Retrieval
- Query for all issues without labels
- Include both open and recently closed issues
- Respect batch size limits
- Handle pagination properly
```

#### Phase 2: Content Analysis & Labeling
```
STEP 3: Issue Content Analysis
- Parse issue title for keywords and patterns
- Analyze issue body for technical indicators
- Check for code snippets and error messages
- Identify user stories and acceptance criteria

STEP 4: Label Assignment Logic
- Apply type labels based on content structure
- Determine priority from urgency indicators
- Assign component labels from technical context
- Add status labels based on issue state
```

---

## Label Detection Patterns

### Type Label Detection
```python
TYPE_PATTERNS = {
    "type/bug": [
        "error", "bug", "broken", "crash", "fails", "not working",
        "exception", "regression", "defect", "issue with"
    ],
    "type/feature": [
        "add", "implement", "create", "new feature", "enhancement",
        "support for", "ability to", "would like", "request"
    ],
    "type/epic": [
        "epic:", "multiple", "system", "complete", "full implementation",
        "checklist:", "- [ ]", "comprises", "includes tasks"
    ],
    "type/task": [
        "update", "modify", "refactor", "change", "improve",
        "optimize", "migrate", "upgrade", "convert"
    ],
    "type/documentation": [
        "docs", "documentation", "readme", "guide", "tutorial",
        "example", "clarify", "explain", "document"
    ],
    "type/question": [
        "how to", "?", "what is", "why does", "can someone",
        "help", "confused", "understanding", "clarification"
    ],
    "type/test": [
        "test", "testing", "coverage", "unit test", "integration test",
        "e2e", "spec", "test case", "QA"
    ]
}
```

### Priority Label Detection
```python
PRIORITY_INDICATORS = {
    "priority/critical": [
        "urgent", "critical", "blocker", "production", "down",
        "security", "vulnerability", "data loss", "breaking change"
    ],
    "priority/high": [
        "important", "high priority", "needed for release", "blocking",
        "customer reported", "affects many users", "performance issue"
    ],
    "priority/medium": [
        "should", "would be nice", "planned", "roadmap",
        "moderate impact", "workaround exists"
    ],
    "priority/low": [
        "nice to have", "someday", "minor", "cosmetic",
        "edge case", "rare", "low impact"
    ]
}
```

### Component Label Detection
```python
COMPONENT_DETECTION = {
    "component/api": [
        "endpoint", "API", "REST", "GraphQL", "route",
        "controller", "middleware", "request", "response"
    ],
    "component/frontend": [
        "UI", "UX", "component", "React", "Vue", "Angular",
        "CSS", "style", "layout", "responsive", "button"
    ],
    "component/database": [
        "database", "DB", "SQL", "query", "migration",
        "schema", "index", "performance", "PostgreSQL", "MongoDB"
    ],
    "component/auth": [
        "authentication", "authorization", "login", "JWT",
        "password", "security", "permissions", "roles", "access"
    ],
    "component/infrastructure": [
        "deployment", "CI/CD", "Docker", "Kubernetes",
        "AWS", "cloud", "environment", "configuration"
    ]
}
```

---

## Intelligent Analysis Rules

### Multi-Label Logic
```markdown
# Label Combination Rules

1. **Bug + Component**: Always assign both type and affected component
   Example: "Login button throws error" → `type/bug` + `component/frontend` + `component/auth`

2. **Feature + Priority**: New features get priority based on impact
   Example: "Add SSO support" → `type/feature` + `priority/high` + `component/auth`

3. **Epic Detection**: Issues with checklists or multiple sub-tasks
   Example: Issue with "- [ ] Task 1\n- [ ] Task 2" → `type/epic` + appropriate priority

4. **Cross-Component**: Issues affecting multiple areas get all relevant labels
   Example: "API returns wrong data in UI" → `component/api` + `component/frontend`
```

### Context-Aware Labeling
```markdown
# Advanced Detection Patterns

1. **Error Message Analysis**:
   - Stack traces → `type/bug` + relevant component
   - HTTP status codes → `component/api`
   - Console errors → `component/frontend`

2. **User Story Format**:
   - "As a..." format → `type/feature` or `type/task`
   - Clear acceptance criteria → higher priority

3. **Technical Debt Indicators**:
   - "refactor", "cleanup", "technical debt" → `type/task` + `priority/medium`
   - "TODO" or "FIXME" references → `type/task`

4. **Security Indicators**:
   - Security-related keywords → `priority/critical` + `component/auth`
   - Vulnerability mentions → `type/bug` + `priority/critical`
```

---

## Command Implementation

### Execution Flow
```yaml
1. Repository Setup:
   - Authenticate with GitHub
   - Verify repository access
   - Load existing labels

2. Issue Processing:
   - Fetch unlabeled issues (respecting batch size)
   - For each issue:
     a. Analyze title and body
     b. Apply pattern matching
     c. Determine label set
     d. Apply labels (or preview in dry-run)
     e. Log decision rationale

3. Label Management:
   - Check if determined labels exist
   - Create missing labels if --create-missing
   - Use sensible color schemes for new labels

4. Reporting:
   - Summary of processed issues
   - Labels applied per category
   - Any errors or skipped issues
   - Recommendations for manual review
```

### Output Format
```markdown
# Auto-Labeling Report

## Summary
- **Total Issues Scanned**: 127
- **Unlabeled Issues Found**: 43
- **Issues Labeled**: 41
- **Issues Skipped**: 2 (insufficient information)
- **New Labels Created**: 3

## Labeling Decisions

### Issue #234: "Login button not working after update"
**Labels Applied**: `type/bug`, `priority/high`, `component/frontend`, `component/auth`
**Rationale**: Detected bug keywords, auth-related, UI component affected

### Issue #235: "Add dark mode support"
**Labels Applied**: `type/feature`, `priority/medium`, `component/frontend`
**Rationale**: Feature request pattern, UI enhancement, no urgency indicators

### Issue #236: "Improve API documentation"
**Labels Applied**: `type/documentation`, `priority/low`, `component/api`
**Rationale**: Documentation keywords, API-related content

[... additional issues ...]

## Manual Review Recommended
- Issue #237: Complex issue mentioning multiple unrelated topics
- Issue #240: Insufficient description for accurate categorization

## New Labels Created
- `component/testing` - Color: #0E8A16
- `type/research` - Color: #D93F0B
- `component/performance` - Color: #006B75
```

---

## Advanced Features

### Dry Run Mode
```bash
/auto-label-issues --dry-run
```
- Shows what labels would be applied without making changes
- Useful for reviewing labeling logic before applying
- Generates full report with proposed changes

### Custom Label Mappings
```yaml
# .github/claude-label-config.yml
custom_patterns:
  "needs-design":
    - "mockup"
    - "wireframe"
    - "user interface"
    - "design review"
  
  "needs-research":
    - "investigate"
    - "explore options"
    - "proof of concept"
    - "spike"

label_aliases:
  "bug": "type/defect"  # Map common terms to repo standards
  "enhancement": "type/improvement"

exclusions:
  - "wontfix"
  - "duplicate"
  - "invalid"
```

### Batch Processing
```bash
/auto-label-issues --batch-size=100 --throttle=2s
```
- Process issues in batches to respect rate limits
- Add delays between API calls
- Resume from last position if interrupted

### Integration with Other Commands
```bash
# First create issues
/create-issues "Authentication system" --epic

# Then auto-label any created without labels
/auto-label-issues --since="1 hour ago"

# Review specific component
/auto-label-issues --filter="component/auth"
```

---

## Error Handling

### Common Scenarios
```markdown
# Insufficient Permissions
❌ Error: Insufficient permissions to label issues.
Required permission: 'issues:write'
Please check your GitHub token permissions.

# Rate Limiting
⚠️ Warning: GitHub API rate limit reached.
Processed 60 of 150 issues. Resuming in 45 minutes...
Use --save-progress to resume from this point.

# Missing Labels
⚠️ Warning: Label 'type/epic' not found in repository.
Use --create-missing flag to automatically create missing labels.
Skipping label application for 5 issues requiring this label.

# Ambiguous Content
ℹ️ Info: Issue #345 has ambiguous content.
Multiple label categories detected with equal confidence.
Applying most specific labels: type/task, priority/medium
Flagged for manual review.
```

---

## Best Practices

### Label Hygiene
1. **Consistent Naming**: Maintain category prefixes (type/, priority/, etc.)
2. **Mutual Exclusivity**: One label per category (one type, one priority)
3. **Color Coding**: Use consistent colors across label categories
4. **Regular Cleanup**: Remove unused labels periodically

### Quality Assurance
1. **Start with Dry Run**: Always preview changes first
2. **Batch Processing**: Process in smaller batches initially
3. **Manual Review**: Check flagged issues manually
4. **Feedback Loop**: Refine patterns based on results

### Repository Standards
1. **Document Labels**: Maintain label descriptions
2. **Team Agreement**: Ensure team consensus on label taxonomy
3. **Automation Rules**: Set up GitHub Actions for label-based automation
4. **Regular Audits**: Schedule periodic label audits

---

## Example Usage Scenarios

### Scenario 1: New Repository Setup
```bash
# Analyze and establish initial labeling
/auto-label-issues --create-missing --dry-run

# Review proposed labels, then apply
/auto-label-issues --create-missing
```

### Scenario 2: Regular Maintenance
```bash
# Weekly label cleanup
/auto-label-issues --since="7 days ago" --batch-size=50
```

### Scenario 3: Post-Migration Cleanup
```bash
# After importing issues from another system
/auto-label-issues --all --create-missing --report=detailed
```

---

*This command intelligently analyzes and labels GitHub issues, maintaining repository organization and improving issue discoverability.*
