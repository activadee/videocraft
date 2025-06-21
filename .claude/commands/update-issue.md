# Claude Code Command: `/update-issue`

## Command Overview
**Purpose**: Update GitHub issues with status, progress, commits, and automated workflow management.

**Syntax**: `/update-issue --issue-number=<number> [--status="<status>"] [--add-commit] [--close] [--assign=<user>] [--label=<label>]`

**Examples**: 
- `/update-issue --issue-number=123 --status="Working on implementation"` - Add status update
- `/update-issue --issue-number=123 --add-commit` - Link latest commit to issue
- `/update-issue --issue-number=123 --close --status="Completed in PR #456"` - Close with final status
- `/update-issue --issue-number=123 --assign=@developer --label=in-progress` - Update assignment and labels

**Note**: Manages issue lifecycle with automated linking to commits and PRs.

---

## Command Description

This command handles GitHub issue updates and workflow management. Claude will:

1. **Fetch** current issue details and context
2. **Update** issue with status, progress, or resolution
3. **Link** commits, PRs, and related work automatically
4. **Manage** labels, assignments, and milestones
5. **Maintain** issue history and communication

---

## System Instructions

### Role Definition
You are a GitHub issue workflow specialist. When the `/update-issue` command is invoked, you will:

1. **Validate the issue** exists and is accessible
2. **Update issue content** with status and progress
3. **Link related work** (commits, PRs, branches)
4. **Manage issue metadata** (labels, assignees, milestones)
5. **Maintain workflow** state and history

### Issue Update Workflow

#### Phase 1: Issue Validation and Context
```bash
# Validate issue exists
if ! gh issue view $ISSUE_NUMBER >/dev/null 2>&1; then
    echo "❌ Issue #$ISSUE_NUMBER not found"
    exit 1
fi

# Get current issue details
ISSUE_DATA=$(gh issue view $ISSUE_NUMBER --json title,body,state,assignees,labels,milestone,comments)
ISSUE_TITLE=$(echo "$ISSUE_DATA" | jq -r .title)
ISSUE_STATE=$(echo "$ISSUE_DATA" | jq -r .state)
CURRENT_ASSIGNEES=$(echo "$ISSUE_DATA" | jq -r '.assignees[].login' | tr '\n' ',' | sed 's/,$//')
CURRENT_LABELS=$(echo "$ISSUE_DATA" | jq -r '.labels[].name' | tr '\n' ',' | sed 's/,$//')

echo "📋 Issue #$ISSUE_NUMBER: $ISSUE_TITLE"
echo "State: $ISSUE_STATE"
echo "Assignees: ${CURRENT_ASSIGNEES:-none}"
echo "Labels: ${CURRENT_LABELS:-none}"
```

#### Phase 2: Status Update
```bash
if [[ -n "$STATUS_MESSAGE" ]]; then
    # Get current branch and commit info
    CURRENT_BRANCH=$(git branch --show-current)
    CURRENT_COMMIT=$(git log --oneline -1)
    
    # Create status update comment
    COMMENT_BODY="## 📊 Status Update

**Status**: $STATUS_MESSAGE
**Branch**: \`$CURRENT_BRANCH\`
**Latest Commit**: $CURRENT_COMMIT
**Updated**: $(date -u +'%Y-%m-%d %H:%M UTC')

---
*Auto-updated via Claude Code*"
    
    # Post comment
    gh issue comment $ISSUE_NUMBER --body "$COMMENT_BODY"
fi
```

#### Phase 3: Commit Linking
```bash
if [[ "$ADD_COMMIT" == "true" ]]; then
    # Get latest commit with issue reference
    LATEST_COMMIT=$(git log --oneline -1)
    COMMIT_SHA=$(echo "$LATEST_COMMIT" | cut -d' ' -f1)
    COMMIT_MSG=$(echo "$LATEST_COMMIT" | cut -d' ' -f2-)
    
    # Check if commit already references issue
    if ! echo "$COMMIT_MSG" | grep -q "#$ISSUE_NUMBER"; then
        # Create commit reference comment
        COMMIT_COMMENT="## 🔗 Related Commit

**Commit**: [\`$COMMIT_SHA\`]($(git remote get-url origin | sed 's/\.git$//')/commit/$COMMIT_SHA)
**Message**: $COMMIT_MSG
**Branch**: \`$CURRENT_BRANCH\`

This commit contributes to resolving this issue.

---
*Auto-linked via Claude Code*"
        
        gh issue comment $ISSUE_NUMBER --body "$COMMIT_COMMENT"
    fi
fi
```

#### Phase 4: Metadata Management
```bash
# Update assignees
if [[ -n "$ASSIGN_USER" ]]; then
    if [[ "$ASSIGN_USER" == "@me" ]]; then
        ASSIGN_USER=$(gh api user --jq .login)
    fi
    
    gh issue edit $ISSUE_NUMBER --add-assignee "$ASSIGN_USER"
    echo "👤 Assigned to $ASSIGN_USER"
fi

# Update labels
if [[ -n "$ADD_LABEL" ]]; then
    gh issue edit $ISSUE_NUMBER --add-label "$ADD_LABEL"
    echo "🏷️ Added label: $ADD_LABEL"
fi

# Close issue if requested
if [[ "$CLOSE_ISSUE" == "true" ]]; then
    CLOSE_REASON="${STATUS_MESSAGE:-Issue resolved}"
    gh issue close $ISSUE_NUMBER --reason completed --comment "## ✅ Issue Resolved

$CLOSE_REASON

---
*Closed via Claude Code*"
    echo "✅ Issue closed"
fi
```

---

## Implementation Steps

### Step 1: Issue Discovery and Validation
```bash
ISSUE_NUMBER="$1"

# Validate issue number provided
if [[ -z "$ISSUE_NUMBER" ]] || [[ ! "$ISSUE_NUMBER" =~ ^[0-9]+$ ]]; then
    echo "❌ Valid issue number required"
    echo "Usage: /update-issue --issue-number=123"
    exit 1
fi

# Check if issue exists and is accessible
echo "🔍 Checking issue #$ISSUE_NUMBER..."

if ! ISSUE_INFO=$(gh issue view $ISSUE_NUMBER --json title,body,state,assignees,labels,milestone,number 2>/dev/null); then
    echo "❌ Issue #$ISSUE_NUMBER not found or not accessible"
    echo "Check issue number and repository permissions"
    exit 1
fi

# Extract issue details
ISSUE_TITLE=$(echo "$ISSUE_INFO" | jq -r .title)
ISSUE_STATE=$(echo "$ISSUE_INFO" | jq -r .state)
ISSUE_ASSIGNEES=$(echo "$ISSUE_INFO" | jq -r '.assignees[]?.login // empty' | tr '\n' ', ' | sed 's/, $//')
ISSUE_LABELS=$(echo "$ISSUE_INFO" | jq -r '.labels[]?.name // empty' | tr '\n' ', ' | sed 's/, $//')

echo "✅ Found issue: $ISSUE_TITLE"
echo "📊 Current state: $ISSUE_STATE"
echo "👥 Assignees: ${ISSUE_ASSIGNEES:-none}"
echo "🏷️ Labels: ${ISSUE_LABELS:-none}"
```

### Step 2: Determine Update Actions
```bash
# Parse command options
ACTIONS=()

if [[ -n "$STATUS_MESSAGE" ]]; then
    ACTIONS+=("status_update")
    echo "📝 Will add status update"
fi

if [[ "$ADD_COMMIT" == "true" ]]; then
    ACTIONS+=("link_commit")
    echo "🔗 Will link latest commit"
fi

if [[ "$CLOSE_ISSUE" == "true" ]]; then
    ACTIONS+=("close_issue")
    echo "✅ Will close issue"
fi

if [[ -n "$ASSIGN_USER" ]]; then
    ACTIONS+=("update_assignee")
    echo "👤 Will update assignee"
fi

if [[ -n "$ADD_LABEL" ]]; then
    ACTIONS+=("add_label")
    echo "🏷️ Will add label"
fi

if [[ ${#ACTIONS[@]} -eq 0 ]]; then
    echo "❌ No actions specified"
    echo "Use --status, --add-commit, --close, --assign, or --label"
    exit 1
fi

echo ""
echo "🚀 Proceeding with ${#ACTIONS[@]} actions..."
```

### Step 3: Execute Status Update
```bash
if [[ " ${ACTIONS[@]} " =~ " status_update " ]]; then
    echo "📝 Adding status update..."
    
    # Get current context
    CURRENT_BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    REPO_URL=$(git remote get-url origin 2>/dev/null | sed 's/\.git$//' || echo "")
    
    # Build status update
    STATUS_UPDATE="## 📊 Status Update

**Status**: $STATUS_MESSAGE

**Context**:
- **Branch**: \`$CURRENT_BRANCH\`
- **Date**: $(date -u +'%Y-%m-%d %H:%M UTC')
- **Reporter**: $(gh api user --jq .login 2>/dev/null || echo "unknown")

"

    # Add commit info if available
    if git log --oneline -1 >/dev/null 2>&1; then
        LATEST_COMMIT=$(git log --oneline -1)
        COMMIT_SHA=$(echo "$LATEST_COMMIT" | cut -d' ' -f1)
        
        STATUS_UPDATE+="**Latest Commit**: [\`$COMMIT_SHA\`]($REPO_URL/commit/$COMMIT_SHA)
"
    fi
    
    # Add PR info if available
    if [[ "$CURRENT_BRANCH" != "main" ]] && [[ "$CURRENT_BRANCH" != "master" ]]; then
        PR_NUMBER=$(gh pr list --head "$CURRENT_BRANCH" --json number --jq '.[0].number // empty' 2>/dev/null)
        if [[ -n "$PR_NUMBER" ]]; then
            STATUS_UPDATE+="**Related PR**: #$PR_NUMBER
"
        fi
    fi
    
    STATUS_UPDATE+="
---
*Auto-updated via Claude Code /update-issue*"
    
    # Post status update
    if gh issue comment $ISSUE_NUMBER --body "$STATUS_UPDATE"; then
        echo "✅ Status update posted"
    else
        echo "❌ Failed to post status update"
        exit 1
    fi
fi
```

### Step 4: Link Commits
```bash
if [[ " ${ACTIONS[@]} " =~ " link_commit " ]]; then
    echo "🔗 Linking latest commit..."
    
    # Get latest commit details
    if ! git log --oneline -1 >/dev/null 2>&1; then
        echo "⚠️ No git commits found"
    else
        LATEST_COMMIT=$(git log --oneline -1)
        COMMIT_SHA=$(echo "$LATEST_COMMIT" | cut -d' ' -f1)
        COMMIT_MSG=$(echo "$LATEST_COMMIT" | cut -d' ' -f2-)
        COMMIT_AUTHOR=$(git log -1 --format='%an')
        COMMIT_DATE=$(git log -1 --format='%ad' --date=short)
        
        # Check if commit already references this issue
        if echo "$COMMIT_MSG" | grep -q "#$ISSUE_NUMBER"; then
            echo "✅ Commit already references issue #$ISSUE_NUMBER"
        else
            # Create commit link comment
            COMMIT_LINK="## 🔗 Related Work

**Commit**: [\`$COMMIT_SHA\`]($REPO_URL/commit/$COMMIT_SHA)
**Message**: $COMMIT_MSG
**Author**: $COMMIT_AUTHOR
**Date**: $COMMIT_DATE
**Branch**: \`$CURRENT_BRANCH\`

This commit contributes to the resolution of this issue.

---
*Auto-linked via Claude Code /update-issue*"
            
            if gh issue comment $ISSUE_NUMBER --body "$COMMIT_LINK"; then
                echo "✅ Commit linked successfully"
            else
                echo "❌ Failed to link commit"
            fi
        fi
    fi
fi
```

### Step 5: Update Metadata
```bash
# Update assignees
if [[ " ${ACTIONS[@]} " =~ " update_assignee " ]]; then
    echo "👤 Updating assignee..."
    
    # Handle special values
    if [[ "$ASSIGN_USER" == "@me" ]]; then
        ASSIGN_USER=$(gh api user --jq .login)
        echo "Assigning to current user: $ASSIGN_USER"
    elif [[ "$ASSIGN_USER" == "none" ]] || [[ "$ASSIGN_USER" == "unassign" ]]; then
        # Remove all assignees
        if gh issue edit $ISSUE_NUMBER --remove-assignee @me 2>/dev/null; then
            echo "✅ Removed assignees"
        fi
    else
        # Add specific assignee
        if gh issue edit $ISSUE_NUMBER --add-assignee "$ASSIGN_USER"; then
            echo "✅ Assigned to $ASSIGN_USER"
        else
            echo "❌ Failed to assign to $ASSIGN_USER"
        fi
    fi
fi

# Add labels
if [[ " ${ACTIONS[@]} " =~ " add_label " ]]; then
    echo "🏷️ Adding label..."
    
    if gh issue edit $ISSUE_NUMBER --add-label "$ADD_LABEL"; then
        echo "✅ Added label: $ADD_LABEL"
    else
        echo "❌ Failed to add label: $ADD_LABEL"
    fi
fi
```

### Step 6: Close Issue (if requested)
```bash
if [[ " ${ACTIONS[@]} " =~ " close_issue " ]]; then
    echo "✅ Closing issue..."
    
    # Prepare closing comment
    CLOSE_COMMENT="## ✅ Issue Resolved

"

    if [[ -n "$STATUS_MESSAGE" ]]; then
        CLOSE_COMMENT+="**Resolution**: $STATUS_MESSAGE
"
    else
        CLOSE_COMMENT+="This issue has been completed.
"
    fi
    
    # Add context
    CLOSE_COMMENT+="
**Context**:
- **Closed by**: $(gh api user --jq .login)
- **Date**: $(date -u +'%Y-%m-%d %H:%M UTC')
- **Branch**: \`$CURRENT_BRANCH\`
"
    
    # Add PR reference if available
    if [[ "$CURRENT_BRANCH" != "main" ]] && [[ "$CURRENT_BRANCH" != "master" ]]; then
        PR_NUMBER=$(gh pr list --head "$CURRENT_BRANCH" --json number --jq '.[0].number // empty' 2>/dev/null)
        if [[ -n "$PR_NUMBER" ]]; then
            CLOSE_COMMENT+="- **Related PR**: #$PR_NUMBER
"
        fi
    fi
    
    CLOSE_COMMENT+="
---
*Issue closed via Claude Code /update-issue*"
    
    # Close the issue
    if gh issue close $ISSUE_NUMBER --reason completed --comment "$CLOSE_COMMENT"; then
        echo "✅ Issue #$ISSUE_NUMBER closed successfully"
    else
        echo "❌ Failed to close issue"
        exit 1
    fi
fi
```

### Step 7: Final Status Report
```bash
echo ""
echo "🎉 Issue update completed!"
echo ""

# Get updated issue info
UPDATED_INFO=$(gh issue view $ISSUE_NUMBER --json title,state,assignees,labels,url)
UPDATED_STATE=$(echo "$UPDATED_INFO" | jq -r .state)
UPDATED_ASSIGNEES=$(echo "$UPDATED_INFO" | jq -r '.assignees[]?.login // empty' | tr '\n' ', ' | sed 's/, $//')
UPDATED_LABELS=$(echo "$UPDATED_INFO" | jq -r '.labels[]?.name // empty' | tr '\n' ', ' | sed 's/, $//')
ISSUE_URL=$(echo "$UPDATED_INFO" | jq -r .url)

echo "📊 Final Status:"
echo "  Issue: #$ISSUE_NUMBER"
echo "  State: $UPDATED_STATE"
echo "  Assignees: ${UPDATED_ASSIGNEES:-none}"
echo "  Labels: ${UPDATED_LABELS:-none}"
echo ""
echo "🔗 Issue URL: $ISSUE_URL"

# Suggest next actions
echo ""
echo "💡 Suggested next steps:"
if [[ "$UPDATED_STATE" == "OPEN" ]]; then
    echo "  • Continue working on the issue"
    echo "  • Use /commit-and-push for code changes"
    echo "  • Use /update-issue --close when resolved"
else
    echo "  • Issue is closed and complete"
    echo "  • Review related PR if applicable"
fi
```

---

## Command Options

### Status Update
```bash
/update-issue --issue-number=123 --status="Implemented user authentication, working on tests"
```
- Add detailed status update with context
- Automatically includes branch, commit, and timestamp information
- Useful for progress tracking and team communication

### Commit Linking
```bash
/update-issue --issue-number=123 --add-commit
```
- Link the latest commit to the issue
- Provides commit details and branch context
- Helps track work progress automatically

### Issue Assignment
```bash
/update-issue --issue-number=123 --assign=@developer
/update-issue --issue-number=123 --assign=@me
/update-issue --issue-number=123 --assign=none
```
- Assign issue to specific user, yourself, or remove assignees
- Supports GitHub usernames and special values
- Updates issue workflow state

### Label Management
```bash
/update-issue --issue-number=123 --label=in-progress
/update-issue --issue-number=123 --label=bug --label=priority-high
```
- Add labels for categorization and workflow
- Supports multiple labels in one command
- Helps with issue organization and filtering

### Close Issue
```bash
/update-issue --issue-number=123 --close --status="Fixed in PR #456"
```
- Close issue with completion status
- Automatically adds closure context and timestamp
- Marks issue as resolved with reason

---

## Error Handling

### Issue Not Found
```
❌ Issue #999 not found or not accessible
Action: Check issue number and repository permissions
```

### Permission Denied
```
❌ Failed to update issue: insufficient permissions
Action: Ensure you have write access to the repository
```

### Invalid User Assignment
```
❌ Failed to assign to unknown-user
Action: Check username exists and has repository access
```

### Invalid Label
```
❌ Failed to add label: invalid-label
Action: Check label exists in repository settings
```

---

## Integration with Other Commands

### Development Workflow
```bash
# Start working on issue
/update-issue --issue-number=123 --status="Starting implementation" --assign=@me --label=in-progress

# Make commits
/commit-and-push "feat: implement user service for issue #123"

# Link work
/update-issue --issue-number=123 --add-commit

# Complete work
/update-issue --issue-number=123 --close --status="Completed in PR #456"
```

### PR Coordination
```bash
# Update issue when creating PR
/update-issue --issue-number=123 --status="Implementation complete, PR created for review"

# Update after PR approval
/update-issue --issue-number=123 --status="PR approved, ready to merge"

# Close when merged
/update-issue --issue-number=123 --close --status="Merged in PR #456"
```

### Team Communication
```bash
# Handoff to reviewer
/update-issue --issue-number=123 --status="Ready for code review" --assign=@reviewer --label=review

# Back to developer
/update-issue --issue-number=123 --status="Addressing review feedback" --assign=@developer --label=in-progress
```

---

## Best Practices

### Status Updates
1. **Be specific**: Describe what was accomplished and what's next
2. **Include context**: Reference commits, PRs, and branches
3. **Update regularly**: Keep stakeholders informed of progress
4. **Use clear language**: Write for team members who may not have context

### Issue Lifecycle
1. **Assign early**: Assign issues when work begins
2. **Track progress**: Link commits and update status regularly
3. **Close properly**: Include resolution details and references
4. **Maintain labels**: Keep categorization current and accurate

### Team Coordination
1. **Communicate changes**: Update when reassigning or changing scope
2. **Reference related work**: Link to PRs, commits, and other issues
3. **Document decisions**: Record important choices in issue comments
4. **Follow conventions**: Use consistent labeling and assignment patterns

---

*This command ensures GitHub issues stay synchronized with development work and team communication.*