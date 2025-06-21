# Claude Code Command: `/update-pr`

## Command Overview
**Purpose**: Update a pull request with current changes, sync with remote, and ensure PR is ready for review.

**Syntax**: `/update-pr [--pr-number=<number>] [--force-push] [--update-description]`

**Examples**: 
- `/update-pr` - Update current PR branch
- `/update-pr --pr-number=123` - Update specific PR
- `/update-pr --force-push` - Force push changes
- `/update-pr --update-description` - Update PR title/description from commits

**Note**: Focuses on PR synchronization and status updates without commits.

---

## Command Description

This command handles PR updates and synchronization. Claude will:

1. **Identify** the target PR (current branch or specified PR number)
2. **Sync** local branch with remote and base branch
3. **Push** any local commits to the PR branch
4. **Update** PR metadata if requested
5. **Display** PR status and next steps

---

## System Instructions

### Role Definition
You are a Git workflow specialist. When the `/update-pr` command is invoked, you will:

1. **Identify the PR** being updated
2. **Sync with remote** to ensure latest changes
3. **Push changes** to the PR branch
4. **Update PR metadata** if requested
5. **Report status** and suggest next actions

### Update Workflow

#### Phase 1: PR Identification
```bash
# Get current branch and PR info
CURRENT_BRANCH=$(git branch --show-current)

if [[ -n "$PR_NUMBER" ]]; then
    # Use specified PR number
    echo "Updating PR #$PR_NUMBER"
    PR_BRANCH=$(gh pr view $PR_NUMBER --json headRefName --jq .headRefName)
    
    # Checkout PR branch if different
    if [[ "$CURRENT_BRANCH" != "$PR_BRANCH" ]]; then
        git checkout $PR_BRANCH
    fi
else
    # Find PR for current branch
    PR_NUMBER=$(gh pr list --head $CURRENT_BRANCH --json number --jq '.[0].number')
    
    if [[ -z "$PR_NUMBER" ]]; then
        echo "No PR found for branch: $CURRENT_BRANCH"
        exit 1
    fi
fi

# Get PR details
gh pr view $PR_NUMBER --json title,body,baseRefName,headRefName,state
```

#### Phase 2: Branch Synchronization
```bash
# Fetch latest changes
git fetch origin

# Get base branch
BASE_BRANCH=$(gh pr view $PR_NUMBER --json baseRefName --jq .baseRefName)

# Check if base branch has new commits
BASE_COMMITS=$(git rev-list --count HEAD..origin/$BASE_BRANCH)
if [[ $BASE_COMMITS -gt 0 ]]; then
    echo "Base branch has $BASE_COMMITS new commits"
    
    # Option to rebase or merge
    echo "Rebasing onto latest $BASE_BRANCH..."
    git rebase origin/$BASE_BRANCH
    
    if [[ $? -ne 0 ]]; then
        echo "Rebase conflicts detected. Please resolve manually."
        exit 1
    fi
fi
```

#### Phase 3: Push Changes
```bash
# Check for unpushed commits
UNPUSHED_COMMITS=$(git rev-list --count origin/$CURRENT_BRANCH..HEAD 2>/dev/null || echo "0")

if [[ $UNPUSHED_COMMITS -gt 0 ]]; then
    echo "Pushing $UNPUSHED_COMMITS commits to PR branch..."
    
    if [[ "$FORCE_PUSH" == "true" ]]; then
        git push origin $CURRENT_BRANCH --force-with-lease
    else
        git push origin $CURRENT_BRANCH
    fi
else
    echo "Branch is up to date with remote"
fi
```

#### Phase 4: Update PR Metadata (Optional)
```bash
if [[ "$UPDATE_DESCRIPTION" == "true" ]]; then
    # Generate new title from recent commits
    RECENT_COMMITS=$(git log origin/$BASE_BRANCH..HEAD --oneline)
    
    # Extract conventional commit types
    COMMIT_TYPES=$(echo "$RECENT_COMMITS" | grep -oE '^[a-f0-9]+ (feat|fix|refactor|docs|test|chore)' | cut -d' ' -f2 | sort -u)
    
    # Generate title based on commit types
    if [[ $(echo "$COMMIT_TYPES" | wc -l) -eq 1 ]]; then
        MAIN_TYPE=$(echo "$COMMIT_TYPES" | head -1)
        NEW_TITLE="$MAIN_TYPE: [generated from commits]"
    else
        NEW_TITLE="feat: multiple improvements [generated from commits]"
    fi
    
    # Update PR title and description
    gh pr edit $PR_NUMBER --title "$NEW_TITLE" --body "$(echo -e 'Auto-generated from commits:\n\n'; git log origin/$BASE_BRANCH..HEAD --oneline)"
fi
```

---

## Implementation Steps

### Step 1: PR Discovery and Validation
```bash
# Identify target PR
if [[ -n "$PR_NUMBER" ]]; then
    # Validate PR exists
    if ! gh pr view $PR_NUMBER >/dev/null 2>&1; then
        echo "PR #$PR_NUMBER not found"
        exit 1
    fi
    
    PR_BRANCH=$(gh pr view $PR_NUMBER --json headRefName --jq .headRefName)
    echo "Updating PR #$PR_NUMBER (branch: $PR_BRANCH)"
else
    # Find PR for current branch
    CURRENT_BRANCH=$(git branch --show-current)
    PR_NUMBER=$(gh pr list --head $CURRENT_BRANCH --json number --jq '.[0].number // empty')
    
    if [[ -z "$PR_NUMBER" ]]; then
        echo "No open PR found for branch: $CURRENT_BRANCH"
        echo "Create a PR first with: gh pr create"
        exit 1
    fi
    
    echo "Found PR #$PR_NUMBER for current branch"
fi
```

### Step 2: Sync with Remote and Base
```bash
# Fetch all remotes
echo "Fetching latest changes..."
git fetch --all

# Get PR info
PR_INFO=$(gh pr view $PR_NUMBER --json baseRefName,headRefName,state)
BASE_BRANCH=$(echo "$PR_INFO" | jq -r .baseRefName)
HEAD_BRANCH=$(echo "$PR_INFO" | jq -r .headRefName)
PR_STATE=$(echo "$PR_INFO" | jq -r .state)

# Check PR state
if [[ "$PR_STATE" != "OPEN" ]]; then
    echo "Warning: PR #$PR_NUMBER is $PR_STATE"
fi

# Ensure we're on the correct branch
if [[ "$(git branch --show-current)" != "$HEAD_BRANCH" ]]; then
    echo "Switching to PR branch: $HEAD_BRANCH"
    git checkout $HEAD_BRANCH
fi

# Check if base branch has updates
git fetch origin $BASE_BRANCH
BEHIND_COUNT=$(git rev-list --count HEAD..origin/$BASE_BRANCH)

if [[ $BEHIND_COUNT -gt 0 ]]; then
    echo "Base branch is $BEHIND_COUNT commits ahead. Rebasing..."
    git rebase origin/$BASE_BRANCH
    
    if [[ $? -ne 0 ]]; then
        echo "❌ Rebase failed with conflicts"
        echo "Resolve conflicts manually and run:"
        echo "  git rebase --continue"
        echo "  git push origin $HEAD_BRANCH --force-with-lease"
        exit 1
    fi
fi
```

### Step 3: Push Updates
```bash
# Check for local commits to push
LOCAL_COMMITS=$(git rev-list --count origin/$HEAD_BRANCH..HEAD 2>/dev/null || echo "0")

if [[ $LOCAL_COMMITS -gt 0 ]]; then
    echo "Pushing $LOCAL_COMMITS local commits..."
    
    # Show commits being pushed
    echo "Commits to push:"
    git log origin/$HEAD_BRANCH..HEAD --oneline
    
    # Push with appropriate force option
    if [[ "$FORCE_PUSH" == "true" ]] || [[ $BEHIND_COUNT -gt 0 ]]; then
        echo "Force pushing (with lease)..."
        git push origin $HEAD_BRANCH --force-with-lease
    else
        echo "Pushing..."
        git push origin $HEAD_BRANCH
    fi
    
    if [[ $? -eq 0 ]]; then
        echo "✅ Successfully pushed to PR #$PR_NUMBER"
    else
        echo "❌ Push failed"
        exit 1
    fi
else
    echo "✅ Branch is up to date with remote"
fi
```

### Step 4: Update PR Metadata (Optional)
```bash
if [[ "$UPDATE_DESCRIPTION" == "true" ]]; then
    echo "Updating PR description from commits..."
    
    # Get commits since base branch
    COMMITS=$(git log origin/$BASE_BRANCH..HEAD --reverse --format="- %s" | head -20)
    
    if [[ -n "$COMMITS" ]]; then
        # Generate description
        NEW_BODY="## Changes

$COMMITS

---
*Auto-updated: $(date)*"
        
        # Update PR
        gh pr edit $PR_NUMBER --body "$NEW_BODY"
        echo "✅ Updated PR description"
    fi
fi
```

### Step 5: Status Report
```bash
# Get final PR status
echo "=== PR Status ==="
gh pr view $PR_NUMBER

# Check CI status
echo "=== CI Status ==="
gh pr checks $PR_NUMBER

# Show what's next
echo "=== Next Steps ==="
echo "• PR updated successfully"
echo "• Review CI checks above"
echo "• PR URL: $(gh pr view $PR_NUMBER --json url --jq .url)"

# Check if PR is ready for review
DRAFT_STATUS=$(gh pr view $PR_NUMBER --json isDraft --jq .isDraft)
if [[ "$DRAFT_STATUS" == "true" ]]; then
    echo "• Convert from draft when ready: gh pr ready $PR_NUMBER"
fi

REVIEW_REQUESTS=$(gh pr view $PR_NUMBER --json reviewRequests --jq '.reviewRequests | length')
if [[ $REVIEW_REQUESTS -eq 0 ]]; then
    echo "• Request reviews when ready: gh pr edit $PR_NUMBER --add-reviewer @reviewer"
fi
```

---

## Command Options

### Specific PR Number
```bash
/update-pr --pr-number=123
```
- Update a specific PR instead of current branch's PR
- Useful when working on multiple PRs or different branches

### Force Push
```bash
/update-pr --force-push
```
- Force push changes (with lease for safety)
- Necessary after rebasing or when remote history diverged
- Uses `--force-with-lease` to prevent overwriting others' work

### Update Description
```bash
/update-pr --update-description
```
- Auto-generate PR description from commit messages
- Useful when commits tell the story better than original description
- Preserves manual description if commits are minimal

---

## Error Handling

### No PR Found
```
❌ No open PR found for branch: feature/new-feature
Action: Create PR first with: gh pr create
```

### Rebase Conflicts
```
❌ Rebase failed with conflicts in: src/file.go
Action: Resolve conflicts manually:
  1. Edit conflicted files
  2. git add <resolved-files>
  3. git rebase --continue
  4. Re-run /update-pr
```

### Push Rejected
```
❌ Push failed - remote has newer commits
Action: Use --force-push flag or pull latest changes:
  /update-pr --force-push
```

### Closed PR
```
⚠️ Warning: PR #123 is MERGED
Action: PR is already closed, no updates needed
```

---

## Integration with Other Commands

### Before Committing
```bash
# Update PR with existing changes
/update-pr

# Then make new commits
/commit-and-push "fix: address review feedback"
```

### After Review Changes
```bash
# Address review comments
/review-changes <comment-url>

# Update PR with fixes
/update-pr --update-description
```

### Issue Updates
```bash
# Update PR
/update-pr

# Update related issue
/update-issue --issue-number=456 --status="PR updated and ready for review"
```

---

## Best Practices

### Regular Synchronization
1. **Daily Updates**: Run before starting work to sync with base branch
2. **Pre-Review**: Update before requesting reviews
3. **Post-Changes**: Update after making requested changes

### Commit Strategy
1. **Clean History**: Rebase instead of merge when possible
2. **Force Push Safety**: Always use `--force-with-lease`
3. **Atomic Updates**: One logical change per update

### Communication
1. **Status Visibility**: Keep PR description current
2. **CI Awareness**: Monitor checks after updates
3. **Review Coordination**: Notify reviewers of significant updates

---

*This command ensures PRs stay synchronized and up-to-date with minimal friction.*