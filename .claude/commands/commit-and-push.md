# Claude Code Command: `/commit-and-push`

## Command Overview
**Purpose**: Create atomic commits with conventional commit messages and push to remote branch.

**Syntax**: `/commit-and-push "<message>" [--amend] [--no-verify] [--force-push]`

**Examples**: 
- `/commit-and-push "fix: resolve null pointer in user service"` - Standard commit
- `/commit-and-push "feat: add user authentication" --amend` - Amend last commit
- `/commit-and-push "test: add integration tests" --no-verify` - Skip pre-commit hooks
- `/commit-and-push "refactor: optimize database queries" --force-push` - Force push after amend

**Note**: Focuses on atomic commits with proper messaging and safe pushing.

---

## Command Description

This command handles atomic commit creation and pushing. Claude will:

1. **Validate** working directory state and staged changes
2. **Create** commit with conventional commit message format
3. **Run** quality checks (linting, testing) unless skipped
4. **Push** changes to remote branch safely
5. **Report** commit status and next steps

---

## System Instructions

### Role Definition
You are a Git commit specialist. When the `/commit-and-push` command is invoked, you will:

1. **Validate the working directory** state
2. **Stage appropriate changes** if needed
3. **Create atomic commits** with descriptive messages
4. **Run quality checks** before pushing
5. **Push safely** to remote branch

### Commit Workflow

#### Phase 1: Pre-Commit Validation
```bash
# Check git status
git status --porcelain

# Ensure we're not on main/master
CURRENT_BRANCH=$(git branch --show-current)
if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" ]]; then
    echo "❌ Cannot commit directly to $CURRENT_BRANCH"
    echo "Switch to a feature branch first"
    exit 1
fi

# Check for staged changes
STAGED_FILES=$(git diff --cached --name-only)
if [[ -z "$STAGED_FILES" ]]; then
    echo "No staged changes found. Staging all modified files..."
    git add .
fi

# Validate commit message format (conventional commits)
MESSAGE="$1"
if ! echo "$MESSAGE" | grep -qE '^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .+'; then
    echo "⚠️ Message doesn't follow conventional commit format"
    echo "Expected: type(scope): description"
    echo "Examples: feat: add user auth, fix(api): resolve timeout"
fi
```

#### Phase 2: Quality Checks
```bash
# Run linting (unless --no-verify)
if [[ "$NO_VERIFY" != "true" ]]; then
    echo "Running pre-commit checks..."
    
    # Detect project type and run appropriate checks
    if [[ -f "Makefile" ]]; then
        echo "Running make lint..."
        make lint
    elif [[ -f "package.json" ]]; then
        echo "Running npm run lint..."
        npm run lint
    elif [[ -f "go.mod" ]]; then
        echo "Running go vet and gofmt..."
        go vet ./...
        gofmt -d .
    fi
    
    # Run tests if available
    if [[ -f "Makefile" ]]; then
        make test
    elif [[ -f "package.json" ]] && grep -q '"test"' package.json; then
        npm test
    elif [[ -f "go.mod" ]]; then
        go test ./...
    fi
fi
```

#### Phase 3: Commit Creation
```bash
# Create commit
if [[ "$AMEND" == "true" ]]; then
    git commit --amend -m "$MESSAGE"
else
    git commit -m "$MESSAGE"
fi

# Show commit details
echo "Created commit:"
git log --oneline -1
git show --stat HEAD
```

#### Phase 4: Push to Remote
```bash
# Get remote info
REMOTE_BRANCH=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo "")

if [[ -z "$REMOTE_BRANCH" ]]; then
    # First push - set upstream
    echo "Setting upstream and pushing..."
    git push -u origin $CURRENT_BRANCH
else
    # Regular push
    if [[ "$FORCE_PUSH" == "true" ]]; then
        echo "Force pushing (with lease)..."
        git push --force-with-lease
    else
        echo "Pushing to remote..."
        git push
    fi
fi
```

---

## Implementation Steps

### Step 1: Workspace Validation
```bash
# Check current branch
CURRENT_BRANCH=$(git branch --show-current)
echo "Current branch: $CURRENT_BRANCH"

# Prevent commits to main branches
if [[ "$CURRENT_BRANCH" == "main" || "$CURRENT_BRANCH" == "master" ]]; then
    echo "❌ Error: Cannot commit directly to $CURRENT_BRANCH"
    echo "Create a feature branch first:"
    echo "  git checkout -b feature/your-feature-name"
    exit 1
fi

# Check working directory status
if ! git diff-index --quiet HEAD --; then
    echo "📁 Working directory has changes"
else
    echo "✅ Working directory is clean"
fi

# Check staging area
STAGED_COUNT=$(git diff --cached --name-only | wc -l)
UNSTAGED_COUNT=$(git diff --name-only | wc -l)

echo "📊 Changes: $STAGED_COUNT staged, $UNSTAGED_COUNT unstaged"
```

### Step 2: Stage Changes Intelligently
```bash
if [[ $STAGED_COUNT -eq 0 ]]; then
    if [[ $UNSTAGED_COUNT -gt 0 ]]; then
        echo "Auto-staging modified files..."
        
        # Show what will be staged
        echo "Files to be staged:"
        git diff --name-only | sed 's/^/  /'
        
        # Stage all changes
        git add .
        
        echo "✅ Staged all changes"
    else
        echo "❌ No changes to commit"
        exit 1
    fi
else
    echo "✅ Using existing staged changes"
    
    # Show staged files
    echo "Staged files:"
    git diff --cached --name-only | sed 's/^/  /'
fi
```

### Step 3: Validate Commit Message
```bash
MESSAGE="$1"

if [[ -z "$MESSAGE" ]]; then
    echo "❌ Commit message is required"
    echo "Usage: /commit-and-push \"<message>\""
    exit 1
fi

# Check conventional commit format
if echo "$MESSAGE" | grep -qE '^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?: .+'; then
    echo "✅ Conventional commit format detected"
    
    # Extract type
    COMMIT_TYPE=$(echo "$MESSAGE" | sed 's/\(.*\):.*/\1/')
    echo "Commit type: $COMMIT_TYPE"
else
    echo "⚠️ Message doesn't follow conventional commit format"
    echo "Expected: type(scope): description"
    echo "Will proceed anyway..."
fi

# Validate message length
if [[ ${#MESSAGE} -gt 72 ]]; then
    echo "⚠️ Commit message is longer than 72 characters"
    echo "Consider shorter summary and use body for details"
fi
```

### Step 4: Run Quality Checks
```bash
if [[ "$NO_VERIFY" != "true" ]]; then
    echo "🔍 Running pre-commit quality checks..."
    
    # Determine project type and run appropriate checks
    PROJECT_TYPE=""
    
    if [[ -f "Makefile" ]]; then
        PROJECT_TYPE="makefile"
        echo "Detected Makefile project"
        
        # Run make targets if they exist
        if make -n lint >/dev/null 2>&1; then
            echo "Running make lint..."
            if ! make lint; then
                echo "❌ Linting failed"
                exit 1
            fi
        fi
        
        if make -n test >/dev/null 2>&1; then
            echo "Running make test..."
            if ! make test; then
                echo "❌ Tests failed"
                exit 1
            fi
        fi
        
    elif [[ -f "package.json" ]]; then
        PROJECT_TYPE="node"
        echo "Detected Node.js project"
        
        # Run npm scripts if they exist
        if npm run lint --silent >/dev/null 2>&1; then
            echo "Running npm run lint..."
            if ! npm run lint; then
                echo "❌ Linting failed"
                exit 1
            fi
        fi
        
        if npm run test --silent >/dev/null 2>&1; then
            echo "Running npm test..."
            if ! npm test; then
                echo "❌ Tests failed"
                exit 1
            fi
        fi
        
    elif [[ -f "go.mod" ]]; then
        PROJECT_TYPE="go"
        echo "Detected Go project"
        
        # Run Go quality checks
        echo "Running go vet..."
        if ! go vet ./...; then
            echo "❌ go vet failed"
            exit 1
        fi
        
        echo "Running go test..."
        if ! go test ./...; then
            echo "❌ Tests failed"
            exit 1
        fi
        
        echo "Checking gofmt..."
        if [[ -n "$(gofmt -l .)" ]]; then
            echo "❌ Code is not properly formatted"
            echo "Run: gofmt -w ."
            exit 1
        fi
    fi
    
    echo "✅ All quality checks passed"
else
    echo "⚠️ Skipping pre-commit hooks (--no-verify)"
fi
```

### Step 5: Create Commit
```bash
echo "📝 Creating commit..."

COMMIT_ARGS=()
COMMIT_ARGS+=("-m" "$MESSAGE")

if [[ "$NO_VERIFY" == "true" ]]; then
    COMMIT_ARGS+=("--no-verify")
fi

if [[ "$AMEND" == "true" ]]; then
    echo "Amending previous commit..."
    COMMIT_ARGS+=("--amend")
fi

# Execute commit
if git commit "${COMMIT_ARGS[@]}"; then
    echo "✅ Commit created successfully"
    
    # Show commit details
    echo ""
    echo "📋 Commit details:"
    git log --oneline -1
    git show --stat HEAD
else
    echo "❌ Commit failed"
    exit 1
fi
```

### Step 6: Push to Remote
```bash
echo "📤 Pushing to remote..."

# Check if remote tracking is set up
UPSTREAM=$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || echo "")

if [[ -z "$UPSTREAM" ]]; then
    echo "Setting up remote tracking..."
    
    # Push and set upstream
    if git push -u origin $CURRENT_BRANCH; then
        echo "✅ Pushed and set upstream to origin/$CURRENT_BRANCH"
    else
        echo "❌ Failed to push to remote"
        exit 1
    fi
    
else
    echo "Remote tracking: $UPSTREAM"
    
    # Check if we need force push
    if [[ "$AMEND" == "true" ]] || [[ "$FORCE_PUSH" == "true" ]]; then
        echo "Force pushing with lease..."
        
        if git push --force-with-lease; then
            echo "✅ Force pushed successfully"
        else
            echo "❌ Force push failed"
            echo "Remote may have newer commits"
            exit 1
        fi
    else
        echo "Pushing..."
        
        if git push; then
            echo "✅ Pushed successfully"
        else
            echo "❌ Push failed"
            echo "Try: git pull --rebase && git push"
            exit 1
        fi
    fi
fi
```

### Step 7: Status Report
```bash
echo ""
echo "🎉 Commit and push completed!"
echo ""
echo "📊 Summary:"
echo "  Branch: $CURRENT_BRANCH"
echo "  Commit: $(git log --oneline -1)"
echo "  Remote: $(git rev-parse --abbrev-ref --symbolic-full-name @{u})"

# Check if this branch has a PR
PR_NUMBER=$(gh pr list --head $CURRENT_BRANCH --json number --jq '.[0].number // empty' 2>/dev/null || echo "")

if [[ -n "$PR_NUMBER" ]]; then
    echo "  PR: #$PR_NUMBER"
    echo ""
    echo "🔗 PR URL: $(gh pr view $PR_NUMBER --json url --jq .url)"
    echo ""
    echo "💡 Next steps:"
    echo "  • Review CI checks in PR"
    echo "  • Request reviews if ready"
    echo "  • Use /update-pr to sync PR status"
else
    echo ""
    echo "💡 Next steps:"
    echo "  • Create PR: gh pr create"
    echo "  • Continue development with more commits"
fi
```

---

## Command Options

### Amend Last Commit
```bash
/commit-and-push "fix: correct typo in documentation" --amend
```
- Modify the last commit instead of creating new one
- Useful for fixing small mistakes or improving commit messages
- Automatically triggers force push with lease

### Skip Pre-commit Hooks
```bash
/commit-and-push "wip: work in progress" --no-verify
```
- Skip linting, testing, and other pre-commit validations
- Useful for work-in-progress commits or emergency fixes
- Should be used sparingly

### Force Push
```bash
/commit-and-push "refactor: restructure codebase" --force-push
```
- Force push changes to remote (with lease for safety)
- Necessary after rebasing or amending commits
- Uses `--force-with-lease` to prevent data loss

---

## Error Handling

### Main Branch Protection
```
❌ Error: Cannot commit directly to main
Action: Create feature branch:
  git checkout -b feature/your-feature-name
```

### Quality Check Failures
```
❌ Linting failed
Error: src/file.go:15: unused variable 'x'
Action: Fix linting errors or use --no-verify
```

### Push Conflicts
```
❌ Push failed - remote has newer commits
Action: Rebase and retry:
  git pull --rebase
  /commit-and-push "<message>"
```

### No Changes
```
❌ No changes to commit
Action: Make changes first or check git status
```

---

## Integration with Other Commands

### After Code Changes
```bash
# Make changes and commit
/commit-and-push "feat: implement user authentication"

# Update PR
/update-pr --update-description
```

### Before Review
```bash
# Address review feedback
/commit-and-push "fix: address security review comments"

# Reply to specific comment
/reply-comment <comment-url> "Fixed in latest commit"
```

### Issue Workflow
```bash
# Commit solution
/commit-and-push "fix: resolve database timeout issue"

# Update issue
/update-issue --issue-number=123 --status="Fixed in commit abc1234"
```

---

## Best Practices

### Commit Message Guidelines
1. **Use conventional commits**: `type(scope): description`
2. **Be specific**: Describe what changed, not what you did
3. **Present tense**: "add feature" not "added feature"
4. **Imperative mood**: "fix bug" not "fixes bug"
5. **Reference issues**: "fix: resolve timeout (closes #123)"

### Commit Atomicity
1. **One logical change**: Each commit should do one thing
2. **Complete changes**: Don't commit broken or incomplete code
3. **Self-contained**: Each commit should be reviewable independently
4. **Reversible**: Should be safe to revert individual commits

### Quality Assurance
1. **Always test**: Ensure tests pass before committing
2. **Lint consistently**: Fix style issues before pushing
3. **Review changes**: Check `git diff` before committing
4. **Meaningful commits**: Avoid "fix typo" or "wip" in main branches

---

*This command ensures clean, atomic commits with proper validation and safe pushing.*