# Claude Code Command: `/review-changes`

## Command Overview
**Purpose**: Address a specific CodeRabbit AI review comment by implementing the suggested fix, testing, and committing changes.

**Syntax**: `/review-changes <comment-url> [--auto-commit] [--test-command=<cmd>] [--decline] [--with-gemini]`

**Examples**: 
- `/review-changes https://github.com/owner/repo/pull/123#issuecomment-456789` - Address specific comment
- `/review-changes <comment-url> --auto-commit` - Auto-commit after successful fix
- `/review-changes <comment-url> --test-command="make test"` - Custom test command
- `/review-changes <comment-url> --with-gemini` - Discuss with Gemini before implementing
- `/review-changes <comment-url> --decline "Reason for declining"` - Decline the suggestion with explanation

**Note**: Focuses on implementation and testing. Use `/reply-comment` separately for comment responses.

---

## Command Description

This command addresses a specific CodeRabbit AI review comment systematically. Claude will:

1. **Parse** the comment URL to extract PR, repo, and comment details
2. **Fetch** the specific comment content and context using GitHub CLI
3. **Checkout** the appropriate branch for the PR
4. **Analyze** the comment and optionally discuss with Gemini for validation
5. **Implement** the suggested fix or document decline reasoning
6. **Test** changes to ensure no regressions (if implementing)
7. **Commit** fixes with descriptive messages (if implementing)

**Note**: Use `/reply-comment <comment-url>` afterward to communicate implementation status to reviewers.

---

## System Instructions

### Role Definition
You are a senior software engineer specializing in code review implementation. When the `/review-changes` command is invoked, you will:

1. **Parse the comment URL** to identify PR and comment details
2. **Fetch comment content** using GitHub CLI
3. **Understand the specific review feedback**
4. **Optionally discuss with Gemini** for validation and alternative approaches
5. **Implement the suggested fix OR decline** with proper justification
6. **Test and commit changes** atomically (if implementing)

**Communication**: After implementation, use `/reply-comment <comment-url>` to notify reviewers of completion.

### Review Response Workflow

#### Phase 1: Comment Analysis
```
STEP 1: URL Parsing and Setup
- Parse comment URL to extract owner, repo, PR number, comment ID
- Use `gh api` to fetch PR details and comment content
- Checkout the PR branch and pull latest changes
- Ensure clean working directory

STEP 2: Comment Understanding
- Parse CodeRabbit comment content and suggestions
- Identify affected file(s) and line numbers
- Understand the specific issue being raised
- Categorize comment type (bug, style, performance, security)
```

#### Phase 2: Analysis and Decision
```
STEP 3: Gemini Collaboration (Optional)
- Discuss the CodeRabbit suggestion with Gemini for validation
- Explore alternative implementation approaches
- Assess potential risks and trade-offs
- Get independent perspective on the suggestion

STEP 4: Implementation Decision
- Decide whether to implement, modify, or decline the suggestion
- If declining: prepare detailed justification
- If implementing: plan the specific approach
- Consider project context and architectural implications
```

#### Phase 3: Implementation or Decline
```
STEP 5A: Implement Fix (if accepting)
- Apply the suggested changes to the codebase
- Follow project coding standards and conventions
- Ensure fix addresses the specific concern raised
- Verify no unintended side effects

STEP 5B: Decline Suggestion (if rejecting)
- Document specific reasons for declining
- Provide alternative solutions if applicable
- Reference project constraints or architectural decisions

STEP 6: Test and Commit (if implementing)
- Run project test suite to ensure no regressions
- Commit changes with descriptive message
- Push changes to the PR branch

STEP 7: Documentation (if declining)
- Document decline reasoning for team reference
- Save rationale for future /reply-comment usage
```

---

## Implementation Steps

### Step 1: Parse Comment URL and Fetch Details
```bash
# Extract details from comment URL
# Format: https://github.com/owner/repo/pull/123#issuecomment-456789
COMMENT_URL="[provided-url]"
OWNER=$(echo $COMMENT_URL | sed 's|https://github.com/\([^/]*\)/.*|\1|')
REPO=$(echo $COMMENT_URL | sed 's|https://github.com/[^/]*/\([^/]*\)/.*|\1|')
PR_NUMBER=$(echo $COMMENT_URL | sed 's|.*/pull/\([0-9]*\)#.*|\1|')
COMMENT_ID=$(echo $COMMENT_URL | sed 's|.*#issuecomment-\([0-9]*\)|\1|')

# Fetch PR and comment details
gh pr view $PR_NUMBER --repo $OWNER/$REPO --json headRefName,baseRefName
gh api repos/$OWNER/$REPO/issues/comments/$COMMENT_ID
```

### Step 2: Checkout PR Branch and Analyze Comment
```bash
# Checkout PR branch
PR_BRANCH=$(gh pr view $PR_NUMBER --repo $OWNER/$REPO --json headRefName --jq .headRefName)
git checkout $PR_BRANCH
git pull origin $PR_BRANCH

# Parse comment content for:
# - Affected file(s) and line numbers
# - Specific suggestion or issue raised
# - Type of change needed (bug fix, style, performance, etc.)
# - Severity and impact assessment
```

### Step 2.5: Gemini Consultation (Optional)
```
# If --with-gemini flag is used, discuss the suggestion with Gemini:

GEMINI_PROMPT="
I'm reviewing a CodeRabbit AI suggestion on a pull request:

**Original Comment**: [CodeRabbit's suggestion]
**Affected File**: [file path and lines]
**Suggestion Type**: [bug/performance/security/style]
**Project Context**: [brief project description]

**Current Code**:
[relevant code section]

**Suggested Change**:
[CodeRabbit's specific suggestion]

Please analyze this suggestion and provide:
1. Is this suggestion technically sound?
2. Are there any potential risks or downsides?
3. Alternative approaches we should consider?
4. Assessment of priority (critical/important/nice-to-have)?
5. Should we implement as-is, modify, or decline?

Focus on technical accuracy, architectural fit, and potential impact.
"

# Get Gemini's analysis and factor it into the implementation decision
```

### Step 3: Decision and Implementation
```bash
# Decision Point: Implement or Decline?

if [[ "$DECLINE_FLAG" == "true" ]]; then
    # Declining the suggestion
    echo "Declining CodeRabbit suggestion with reason: $DECLINE_REASON"
    # Proceed directly to Step 4 (Reply) with decline explanation
    
else
    # Implementing the suggestion
    echo "Implementing CodeRabbit suggestion..."
    
    # Apply the suggested changes
    # Follow the specific guidance in the CodeRabbit comment
    # Consider Gemini's input if consultation was performed
    # Ensure changes align with project conventions
    
    # Run tests to verify no regressions
    TEST_COMMAND="${TEST_COMMAND:-make test}"
    $TEST_COMMAND
    
    # If tests pass, proceed to commit
    if [[ $? -eq 0 ]]; then
        echo "Tests passed, proceeding to commit"
    else
        echo "Tests failed, reverting changes"
        git checkout -- .
        exit 1
    fi
fi
```

### Step 4: Commit Changes or Document Decline
```bash
if [[ "$DECLINE_FLAG" == "true" ]]; then
    # Document decline reasoning
    echo "📝 Documenting decline reasoning..."
    
    DECLINE_DOC="DECLINE_REASON: $DECLINE_REASON
COMMENT_URL: $COMMENT_URL
COMMENT_ID: $COMMENT_ID
DECLINED_AT: $(date -u +'%Y-%m-%d %H:%M UTC')
GEMINI_CONSULTED: ${GEMINI_CONSULTED:-false}

Use /reply-comment $COMMENT_URL with the decline reasoning to communicate with reviewers."
    
    echo "$DECLINE_DOC" > .claude/decline-$(basename $COMMENT_ID).txt
    echo "✅ Decline reasoning saved for future reference"
    echo "💬 Use: /reply-comment $COMMENT_URL --template=declined"

else
    # Commit changes with descriptive message
    echo "📝 Committing implementation..."
    
    git add .
    git commit -m "fix: address CodeRabbit review comment

Implemented suggestion from comment $COMMENT_ID:
[brief description of change]

$(if [[ "$GEMINI_CONSULTED" == "true" ]]; then echo "Validated with Gemini AI for technical accuracy"; fi)

Fixes: $COMMENT_URL"

    # Push changes
    git push origin $PR_BRANCH
    
    echo "✅ Changes committed and pushed"
    echo "💬 Use: /reply-comment $COMMENT_URL \"Implemented as suggested\" --with-commit"
fi
```

---

## Comment Analysis and Response

### Comment Type Identification
Analyze the CodeRabbit comment to determine the type of issue:

- **Bug/Error**: Issues with logic, null references, incorrect behavior
- **Performance**: Optimization opportunities, inefficient algorithms, resource usage
- **Security**: Vulnerabilities, unsafe practices, data exposure risks
- **Style/Convention**: Code formatting, naming conventions, consistency
- **Refactoring**: Code structure improvements, maintainability enhancements

### Response Strategy by Type

**Bug Comments**: 
- Implement defensive coding and proper error handling
- Add tests for edge cases and error scenarios
- Verify fix resolves the specific issue mentioned

**Performance Comments**:
- Optimize algorithms or reduce computational complexity
- Consider caching, lazy loading, or more efficient data structures
- Measure performance before and after if significant

**Security Comments**:
- Implement security best practices immediately
- Add input validation, sanitization, or proper authentication
- Consider security implications of the change

**Style Comments**:
- Apply project style guide and linting rules
- Ensure consistency with existing codebase patterns
- Use project-specific formatters and conventions

---

## GitHub CLI Integration

### Required GitHub CLI Commands
```bash
# Fetch PR information
gh pr view <pr-number> --repo <owner>/<repo> --json headRefName,baseRefName,title,body

# Fetch specific comment
gh api repos/<owner>/<repo>/issues/comments/<comment-id>

# Reply to comment
gh api repos/<owner>/<repo>/issues/comments/<comment-id>/replies \
  --method POST \
  --field body="<reply-content>"

# Alternative: Create new comment on PR
gh pr comment <pr-number> --repo <owner>/<repo> --body "<reply-content>"
```

### Execution Flow
```yaml
1. Parse Comment URL:
   - Extract owner, repo, PR number, comment ID
   - Validate URL format and accessibility

2. Fetch Context:
   - Get PR details (branch names, title)
   - Fetch specific comment content
   - Identify affected files and lines

3. Implement Fix:
   - Checkout PR branch and update
   - Apply suggested changes from comment
   - Ensure adherence to project conventions

4. Test and Commit:
   - Run test suite to verify no regressions
   - Commit with descriptive message linking to comment
   - Push changes to PR branch

5. Respond:
   - Reply to original comment with @coderabbitai mention
   - Include commit link and brief description of changes
   - Confirm issue resolution
```

---

## Command Options

### Auto-Commit Mode
```bash
/review-changes <comment-url> --auto-commit
```
- Automatically commits and pushes after successful fix and tests
- Uses descriptive commit message with comment reference
- Skips manual confirmation steps
- Ready for follow-up with /reply-comment

### Custom Test Command
```bash
/review-changes <comment-url> --test-command="make test && npm run lint"
```
- Override default test command for project
- Support complex test pipelines
- Ensure all quality checks pass before commit
- Custom validation specific to the change type

### Gemini Consultation Mode
```bash
/review-changes <comment-url> --with-gemini
```
- Discuss the CodeRabbit suggestion with Gemini AI before implementation
- Get independent technical validation of the suggestion
- Explore alternative approaches and potential risks
- Make more informed decisions on complex suggestions

### Decline Suggestion
```bash
/review-changes <comment-url> --decline "The suggested approach conflicts with our architecture pattern X"
```
- Politely decline the CodeRabbit suggestion with clear reasoning
- Provide technical justification for the decision
- Maintain positive communication with the AI reviewer
- Document the decision for future reference

---

## Execution Progress

### Real-time Status Updates
```
🔧 Addressing CodeRabbit Comment

📍 **Comment**: https://github.com/owner/repo/pull/123#issuecomment-456789
📁 **File**: src/components/UserProfile.tsx:45-52
🎯 **Type**: Performance optimization

**Progress:**
✅ Comment URL parsed successfully
✅ PR details fetched (PR #123)
✅ Branch checked out (feature/user-profile-update)
✅ Comment content analyzed
🤖 Consulting with Gemini AI...
✅ Gemini validation completed - suggestion confirmed
🔧 Implementing suggested optimization...
⏳ Running tests...
⏳ Committing changes...
⏳ Replying to comment...

**Status**: In progress - Testing changes
```

### Status with Decline Decision
```
❌ Declining CodeRabbit Comment

📍 **Comment**: https://github.com/owner/repo/pull/123#issuecomment-789012
📁 **File**: src/auth/security.ts:23-30
🎯 **Type**: Security suggestion

**Analysis:**
✅ Comment URL parsed successfully
✅ PR details fetched (PR #123)
✅ Comment content analyzed
🤖 Consulted with Gemini AI for validation
⚠️ Identified conflicts with existing security architecture

**Decision**: DECLINE
**Reason**: Suggestion conflicts with our established OAuth2 flow patterns

**Status**: Preparing decline response with technical justification
```

### Completion Summary - Implementation
```markdown
✅ **CodeRabbit Comment Addressed Successfully**

**Comment**: #issuecomment-456789
**Fix Applied**: Optimized user data fetching using React.useMemo
**Commit**: [`abc1234`](https://github.com/owner/repo/commit/abc1234)
**Tests**: All passing (45/45)
**Gemini Validation**: ✅ Confirmed technically sound
**Changes Made:**
- Added useMemo hook to prevent unnecessary re-calculations
- Reduced component re-renders by 60%
- Maintained existing functionality

**Next Steps:** Use /reply-comment to notify @coderabbitai of implementation.
```

### Completion Summary - Declined
```markdown
❌ **CodeRabbit Comment Declined with Justification**

**Comment**: #issuecomment-789012
**Suggestion**: Refactor authentication middleware approach
**Decision**: DECLINED
**Reason**: Conflicts with established OAuth2 architecture patterns
**Gemini Consultation**: ✅ Confirmed our assessment
**Analysis Summary:**
- Suggestion would break existing security flow
- Alternative approach already implements similar optimization
- Current implementation follows industry best practices

**Next Steps:** Use /reply-comment with decline reasoning to communicate decision.
```

---

## Default Behavior

### Standard Test Command
- **Go projects**: `make test` (if Makefile exists) or `go test ./...`
- **Node.js projects**: `npm test` (if package.json exists)
- **Python projects**: `pytest` (if pytest.ini exists) or `python -m unittest`
- **Mixed projects**: Run tests for affected language/framework

### Commit Message Format
```
fix: address CodeRabbit review comment

[Brief description of the change made]

Implemented suggestion from CodeRabbit review:
- [Specific change details]

Fixes: [comment-url]
```

---

## Error Handling

### Common Error Scenarios

**Invalid Comment URL**
```
❌ Error: Unable to parse comment URL or access comment.
Possible causes:
- Invalid URL format
- Comment deleted or private repository
- Insufficient GitHub permissions
Action: Verify URL and GitHub authentication.
```

**Test Failures After Fix**
```
❌ Error: Tests failed after implementing suggested fix.
Test Output: [specific failure details]
Action: Reverting changes and investigating issue.
Next: Use /reply-comment to request clarification from reviewer.
```

**Merge Conflicts**
```
❌ Error: Unable to apply changes due to merge conflicts.
Conflict in: [file path]
Action: Manual intervention required.
Suggestion: Resolve conflicts manually and re-run command.
```

**Ambiguous Comment**
```
⚠️ Warning: Comment appears to be informational only.
No specific actionable suggestion found.
Action: Use /reply-comment to ask for clarification on specific changes needed.
```

### Recovery Actions
- **Git State**: Automatically stash uncommitted changes before starting
- **Test Failures**: Revert changes and use /reply-comment to explain issue
- **API Issues**: Retry with exponential backoff, inform user of delays
- **Unclear Comments**: Use /reply-comment to ask for clarification rather than guessing

---

## Usage Examples

### Basic Usage
```bash
# Address a specific CodeRabbit comment
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-456789

# With auto-commit enabled
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-456789 --auto-commit

# With custom test command
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-456789 --test-command="npm run test:ci"

# With Gemini consultation for validation
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-456789 --with-gemini

# Decline a suggestion with reasoning
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-789012 --decline "This approach conflicts with our microservices architecture pattern"

# Complex workflow with Gemini validation and auto-commit
/review-changes https://github.com/myorg/myproject/pull/123#issuecomment-456789 --with-gemini --auto-commit
```

### Workflow Integration
```bash
# Address review comment with implementation
/review-changes <comment-url> --with-gemini

# Communicate implementation to reviewer
/reply-comment <comment-url> "Implemented as suggested" --with-commit

# Update PR with all changes
/update-pr --update-description

# For complex architectural changes that need planning
/create-issues "Refactor authentication system based on CodeRabbit feedback"
```

---

## Best Practices

### Implementation Guidelines
1. **Understand First**: Read the comment carefully and understand the concern
2. **Validate When Needed**: Use Gemini consultation for complex or uncertain suggestions
3. **Decide Thoughtfully**: It's okay to decline suggestions that don't fit the project
4. **Follow or Adapt**: Implement exactly as suggested, or adapt with clear reasoning
5. **Test Thoroughly**: Ensure changes don't break existing functionality
6. **Communicate Clearly**: Explain what was done and reference commits or decline reasons

### Commit Standards
1. **Atomic Changes**: One comment fix per commit
2. **Descriptive Messages**: Clearly describe what was changed and why
3. **Link References**: Include comment URL in commit message
4. **Test Verification**: Only commit when tests pass

### Communication Workflow
1. **Implement First**: Focus on code changes and testing
2. **Document Decisions**: Record implementation or decline reasoning
3. **Use Reply Command**: Follow up with /reply-comment for communication
4. **Be Professional**: Maintain courteous and technical communication
5. **Reference Work**: Link commits and provide specific details

---

## Advanced Features

### Gemini AI Integration
The command integrates with Gemini AI through the MCP multi-AI collaboration tools to provide independent validation of CodeRabbit suggestions:

```
mcp__multi-ai-collab__ask_gemini
mcp__multi-ai-collab__gemini_code_review  
mcp__multi-ai-collab__gemini_think_deep
```

**Benefits of Gemini Consultation:**
- Independent technical validation of suggestions
- Identification of potential risks or architectural conflicts
- Alternative implementation approaches
- Confidence in complex technical decisions
- Reduced false positive implementations

### Decision Framework
When evaluating CodeRabbit suggestions, consider:

**Implement When:**
- Suggestion aligns with project architecture
- Clear technical improvement with low risk
- Fixes genuine bugs or security issues
- Improves code quality without side effects

**Consult Gemini When:**
- Complex architectural changes suggested
- Uncertain about potential side effects
- Suggestion conflicts with existing patterns
- High-impact performance or security changes

**Decline When:**
- Conflicts with established architecture decisions
- Introduces unnecessary complexity
- Breaks existing functionality patterns
- Goes against project-specific constraints

---

*This command focuses on implementing CodeRabbit AI review suggestions with technical validation. Use `/reply-comment` separately for reviewer communication.*