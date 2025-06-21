# Claude Code Command: `/reply-comment`

## Command Overview
**Purpose**: Reply to specific GitHub PR comments (especially CodeRabbit) with implementation status, explanations, or acknowledgments.

**Syntax**: `/reply-comment <comment-url> "<reply-message>" [--with-commit] [--with-gemini] [--template=<type>]`

**Examples**: 
- `/reply-comment https://github.com/owner/repo/pull/123#issuecomment-456789 "Fixed in latest commit"` - Simple reply
- `/reply-comment <comment-url> "Implemented as suggested" --with-commit` - Reply with commit reference
- `/reply-comment <comment-url> "Need clarification on approach" --with-gemini` - Get Gemini input for complex responses
- `/reply-comment <comment-url> --template=acknowledge` - Use acknowledgment template

**Note**: Focused on comment communication without code changes.

---

## Command Description

This command handles GitHub comment replies and communication. Claude will:

1. **Parse** the comment URL to extract PR and comment details
2. **Fetch** the original comment content and context
3. **Compose** appropriate reply with optional AI assistance
4. **Reference** commits, PRs, or other relevant context
5. **Post** professional reply mentioning relevant parties

---

## System Instructions

### Role Definition
You are a code review communication specialist. When the `/reply-comment` command is invoked, you will:

1. **Parse comment URL** and extract context
2. **Fetch original comment** content and thread
3. **Compose appropriate reply** based on context and message
4. **Include relevant references** (commits, PRs, documentation)
5. **Post professional response** maintaining positive communication

### Reply Workflow

#### Phase 1: Comment Analysis
```bash
# Parse comment URL
COMMENT_URL="$1"
REPLY_MESSAGE="$2"

# Extract details from URL (GitHub format)
if [[ "$COMMENT_URL" =~ github\.com/([^/]+)/([^/]+)/pull/([0-9]+)#issuecomment-([0-9]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
    PR_NUMBER="${BASH_REMATCH[3]}"
    COMMENT_ID="${BASH_REMATCH[4]}"
else
    echo "❌ Invalid GitHub comment URL format"
    exit 1
fi

# Fetch comment details
COMMENT_DATA=$(gh api repos/$OWNER/$REPO/issues/comments/$COMMENT_ID)
COMMENT_BODY=$(echo "$COMMENT_DATA" | jq -r .body)
COMMENT_AUTHOR=$(echo "$COMMENT_DATA" | jq -r .user.login)
```

#### Phase 2: Reply Composition
```bash
# Build reply with context
REPLY_BODY="$REPLY_MESSAGE"

# Add commit reference if requested
if [[ "$WITH_COMMIT" == "true" ]]; then
    LATEST_COMMIT=$(git log --oneline -1)
    COMMIT_SHA=$(echo "$LATEST_COMMIT" | cut -d' ' -f1)
    REPO_URL=$(git remote get-url origin | sed 's/\.git$//')
    
    REPLY_BODY+="

**Related commit**: [\`$COMMIT_SHA\`]($REPO_URL/commit/$COMMIT_SHA)"
fi

# Add professional footer
REPLY_BODY+="

---
*Replied via Claude Code*"
```

#### Phase 3: Post Reply
```bash
# Post comment reply
gh api repos/$OWNER/$REPO/issues/comments \
  --method POST \
  --field body="$REPLY_BODY"
```

---

## Implementation Steps

### Step 1: URL Parsing and Validation
```bash
COMMENT_URL="$1"
REPLY_MESSAGE="$2"

# Validate inputs
if [[ -z "$COMMENT_URL" ]]; then
    echo "❌ Comment URL is required"
    echo "Usage: /reply-comment <comment-url> \"<message>\""
    exit 1
fi

if [[ -z "$REPLY_MESSAGE" ]] && [[ -z "$TEMPLATE" ]]; then
    echo "❌ Reply message or template is required"
    echo "Usage: /reply-comment <comment-url> \"<message>\""
    exit 1
fi

echo "🔍 Parsing comment URL..."

# Extract GitHub comment details
if [[ "$COMMENT_URL" =~ https://github\.com/([^/]+)/([^/]+)/pull/([0-9]+)#issuecomment-([0-9]+) ]]; then
    OWNER="${BASH_REMATCH[1]}"
    REPO="${BASH_REMATCH[2]}"
    PR_NUMBER="${BASH_REMATCH[3]}"
    COMMENT_ID="${BASH_REMATCH[4]}"
    
    echo "✅ Parsed GitHub comment:"
    echo "  Owner: $OWNER"
    echo "  Repo: $REPO"
    echo "  PR: #$PR_NUMBER"
    echo "  Comment ID: $COMMENT_ID"
else
    echo "❌ Invalid GitHub comment URL format"
    echo "Expected: https://github.com/owner/repo/pull/123#issuecomment-456789"
    exit 1
fi
```

### Step 2: Fetch Comment Context
```bash
echo "📥 Fetching comment details..."

# Get comment data
if ! COMMENT_DATA=$(gh api repos/$OWNER/$REPO/issues/comments/$COMMENT_ID 2>/dev/null); then
    echo "❌ Failed to fetch comment #$COMMENT_ID"
    echo "Check permissions and comment exists"
    exit 1
fi

# Extract comment details
COMMENT_BODY=$(echo "$COMMENT_DATA" | jq -r .body)
COMMENT_AUTHOR=$(echo "$COMMENT_DATA" | jq -r .user.login)
COMMENT_DATE=$(echo "$COMMENT_DATA" | jq -r .created_at)
COMMENT_URL_API=$(echo "$COMMENT_DATA" | jq -r .html_url)

echo "✅ Found comment by @$COMMENT_AUTHOR"
echo "📅 Created: $COMMENT_DATE"

# Get PR context
PR_DATA=$(gh pr view $PR_NUMBER --repo $OWNER/$REPO --json title,headRefName,baseRefName)
PR_TITLE=$(echo "$PR_DATA" | jq -r .title)
PR_BRANCH=$(echo "$PR_DATA" | jq -r .headRefName)

echo "📋 PR: $PR_TITLE"
echo "🌿 Branch: $PR_BRANCH"

# Show comment preview (first 100 chars)
COMMENT_PREVIEW=$(echo "$COMMENT_BODY" | head -c 100 | tr '\n' ' ')
echo "💬 Comment preview: $COMMENT_PREVIEW..."
```

### Step 3: Generate Reply Content
```bash
echo "✍️ Composing reply..."

# Handle template-based replies
if [[ -n "$TEMPLATE" ]]; then
    case "$TEMPLATE" in
        "acknowledge")
            REPLY_MESSAGE="Thank you for the review! I'll address this suggestion."
            ;;
        "implemented")
            REPLY_MESSAGE="✅ Implemented as suggested. Please review the changes."
            ;;
        "clarification")
            REPLY_MESSAGE="Could you provide more details on the preferred approach for this change?"
            ;;
        "alternative")
            REPLY_MESSAGE="I've implemented a slightly different approach that achieves the same goal. Please let me know if this works for you."
            ;;
        "declined")
            REPLY_MESSAGE="Thank you for the suggestion. After consideration, I believe the current approach is better suited for our use case."
            ;;
        *)
            echo "❌ Unknown template: $TEMPLATE"
            echo "Available templates: acknowledge, implemented, clarification, alternative, declined"
            exit 1
            ;;
    esac
    
    echo "📝 Using template: $TEMPLATE"
fi

# Build base reply
REPLY_BODY="$REPLY_MESSAGE"

# Add mention for original comment author if it's a bot
if [[ "$COMMENT_AUTHOR" == "coderabbitai" ]] || [[ "$COMMENT_AUTHOR" =~ bot$ ]]; then
    REPLY_BODY="@$COMMENT_AUTHOR $REPLY_BODY"
fi
```

### Step 4: Add Context and References
```bash
# Add commit reference if requested
if [[ "$WITH_COMMIT" == "true" ]]; then
    echo "🔗 Adding commit reference..."
    
    if git log --oneline -1 >/dev/null 2>&1; then
        LATEST_COMMIT=$(git log --oneline -1)
        COMMIT_SHA=$(echo "$LATEST_COMMIT" | cut -d' ' -f1)
        COMMIT_MSG=$(echo "$LATEST_COMMIT" | cut -d' ' -f2-)
        
        # Get repository URL
        REPO_URL=$(git remote get-url origin | sed 's/\.git$//' | sed 's/git@github.com:/https:\/\/github.com\//')
        
        REPLY_BODY+="

**Related commit**: [\`$COMMIT_SHA\`]($REPO_URL/commit/$COMMIT_SHA) - $COMMIT_MSG"
        
        echo "✅ Added commit reference: $COMMIT_SHA"
    else
        echo "⚠️ No git commits found, skipping commit reference"
    fi
fi

# Add branch information if relevant
CURRENT_BRANCH=$(git branch --show-current 2>/dev/null)
if [[ -n "$CURRENT_BRANCH" ]] && [[ "$CURRENT_BRANCH" != "main" ]] && [[ "$CURRENT_BRANCH" != "master" ]]; then
    REPLY_BODY+="
**Branch**: \`$CURRENT_BRANCH\`"
fi

# Add timestamp
REPLY_BODY+="

---
*Replied via Claude Code at $(date -u +'%Y-%m-%d %H:%M UTC')*"
```

### Step 5: Gemini Consultation (Optional)
```bash
if [[ "$WITH_GEMINI" == "true" ]]; then
    echo "🤖 Consulting with Gemini for response optimization..."
    
    # Prepare context for Gemini
    GEMINI_PROMPT="I'm replying to a code review comment. Please help optimize my response for professionalism and clarity.

**Original Comment** (by @$COMMENT_AUTHOR):
$COMMENT_BODY

**My Draft Reply**:
$REPLY_MESSAGE

**Context**:
- PR: $PR_TITLE
- This is a GitHub pull request comment thread
- Comment author is: $COMMENT_AUTHOR

Please suggest improvements to make the reply:
1. More professional and courteous
2. Technically clear and specific
3. Appropriately detailed for the context
4. Following good code review communication practices

Provide the improved reply message only, without explanations."
    
    # Get Gemini's suggestion
    if GEMINI_RESPONSE=$(mcp__multi-ai-collab__ask_gemini "$GEMINI_PROMPT" 0.3 2>/dev/null); then
        echo "✅ Received Gemini suggestions"
        echo "🤖 Gemini-optimized reply:"
        echo "$GEMINI_RESPONSE"
        echo ""
        echo "Use Gemini's version? (y/n)"
        read -r USE_GEMINI
        
        if [[ "$USE_GEMINI" =~ ^[Yy] ]]; then
            REPLY_MESSAGE="$GEMINI_RESPONSE"
            REPLY_BODY="@$COMMENT_AUTHOR $REPLY_MESSAGE"
            
            # Re-add context (commit, timestamp) to Gemini version
            if [[ "$WITH_COMMIT" == "true" ]] && [[ -n "$COMMIT_SHA" ]]; then
                REPLY_BODY+="

**Related commit**: [\`$COMMIT_SHA\`]($REPO_URL/commit/$COMMIT_SHA) - $COMMIT_MSG"
            fi
            
            REPLY_BODY+="

---
*Optimized with Gemini AI and posted via Claude Code*"
            
            echo "✅ Using Gemini-optimized reply"
        else
            echo "📝 Using original reply"
        fi
    else
        echo "⚠️ Failed to get Gemini suggestions, using original reply"
    fi
fi
```

### Step 6: Post Reply
```bash
echo "📤 Posting reply to comment..."

# Show final reply preview
echo ""
echo "📋 Final reply:"
echo "----------------------------------------"
echo "$REPLY_BODY"
echo "----------------------------------------"
echo ""

# Post the reply
if gh api repos/$OWNER/$REPO/issues/comments \
    --method POST \
    --field body="$REPLY_BODY" >/dev/null; then
    
    echo "✅ Reply posted successfully"
    
    # Get the new comment ID from the response
    echo "🔗 Comment URL: $COMMENT_URL_API"
    
else
    echo "❌ Failed to post reply"
    echo "Check permissions and try again"
    exit 1
fi
```

### Step 7: Completion Summary
```bash
echo ""
echo "🎉 Comment reply completed!"
echo ""
echo "📊 Summary:"
echo "  Original comment: by @$COMMENT_AUTHOR"
echo "  PR: #$PR_NUMBER - $PR_TITLE"
echo "  Reply posted: $(date -u +'%Y-%m-%d %H:%M UTC')"

if [[ "$WITH_COMMIT" == "true" ]]; then
    echo "  Commit referenced: $COMMIT_SHA"
fi

if [[ "$WITH_GEMINI" == "true" ]]; then
    echo "  Gemini optimization: used"
fi

echo ""
echo "💡 Next steps:"
echo "  • Monitor for follow-up responses"
echo "  • Continue addressing other review comments"
echo "  • Use /update-pr when ready for re-review"
```

---

## Command Options

### Simple Reply
```bash
/reply-comment https://github.com/owner/repo/pull/123#issuecomment-456789 "Thank you for the feedback, I'll address this shortly."
```
- Basic reply to any comment
- Professional and courteous communication
- Suitable for acknowledgments and updates

### Reply with Commit Reference
```bash
/reply-comment <comment-url> "Fixed as suggested" --with-commit
```
- Automatically includes latest commit in reply
- Links to commit with SHA and message
- Useful when replying after implementing changes

### Gemini-Optimized Reply
```bash
/reply-comment <comment-url> "I think the current approach is better" --with-gemini
```
- Get AI assistance for professional communication
- Optimizes tone, clarity, and technical accuracy
- Helpful for complex or sensitive responses

### Template-Based Replies
```bash
/reply-comment <comment-url> --template=acknowledge
/reply-comment <comment-url> --template=implemented
/reply-comment <comment-url> --template=clarification
/reply-comment <comment-url> --template=alternative
/reply-comment <comment-url> --template=declined
```
- Pre-defined professional response templates
- Consistent communication style
- Quick responses for common scenarios

---

## Reply Templates

### Acknowledgment Template
```markdown
Thank you for the review! I'll address this suggestion.
```

### Implementation Template
```markdown
✅ Implemented as suggested. Please review the changes.
```

### Clarification Template  
```markdown
Could you provide more details on the preferred approach for this change?
```

### Alternative Approach Template
```markdown
I've implemented a slightly different approach that achieves the same goal. Please let me know if this works for you.
```

### Declined Suggestion Template
```markdown
Thank you for the suggestion. After consideration, I believe the current approach is better suited for our use case.
```

---

## Error Handling

### Invalid URL Format
```
❌ Invalid GitHub comment URL format
Expected: https://github.com/owner/repo/pull/123#issuecomment-456789
```

### Comment Not Found
```
❌ Failed to fetch comment #456789
Action: Check permissions and verify comment exists
```

### Permission Denied
```
❌ Failed to post reply
Action: Ensure you have write access to the repository
```

### Network Issues
```
❌ Failed to connect to GitHub API
Action: Check internet connection and GitHub status
```

---

## Integration with Other Commands

### Code Review Workflow
```bash
# Address the suggestion with code changes
/commit-and-push "fix: address review feedback on error handling"

# Reply to original comment with implementation
/reply-comment <comment-url> "Implemented error handling as suggested" --with-commit

# Update PR status
/update-pr --update-description
```

### Multi-Comment Responses
```bash
# Reply to multiple comments in sequence
/reply-comment <comment-url-1> "Fixed null check issue" --with-commit
/reply-comment <comment-url-2> "Added unit tests for this function" --with-commit
/reply-comment <comment-url-3> "Clarified variable naming as suggested"

# Final PR update
/update-pr --update-description
```

### Complex Review Discussions
```bash
# Use Gemini for complex technical responses
/reply-comment <comment-url> "The performance implications need consideration" --with-gemini

# Follow up with implementation
/commit-and-push "perf: optimize database query based on review feedback"
/reply-comment <comment-url> "Implemented optimization as discussed" --with-commit
```

---

## Best Practices

### Communication Guidelines
1. **Be professional**: Maintain courteous tone even when disagreeing
2. **Be specific**: Reference exact changes, commits, or line numbers
3. **Be timely**: Respond to comments promptly to maintain momentum
4. **Be grateful**: Thank reviewers for their time and feedback

### Technical Responses
1. **Include context**: Reference commits, PRs, or documentation
2. **Explain reasoning**: When declining suggestions, provide technical justification
3. **Offer alternatives**: When suggesting different approaches, explain benefits
4. **Test references**: Include test results or benchmarks when relevant

### Workflow Integration
1. **Reply after implementing**: Use --with-commit to show work completed
2. **Update status**: Keep team informed of progress and decisions
3. **Maintain threads**: Keep related discussions in the same comment thread
4. **Document decisions**: Record important architectural or design choices

---

*This command ensures professional and effective communication in code review processes.*