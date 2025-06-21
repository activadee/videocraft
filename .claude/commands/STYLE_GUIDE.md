# Claude Code Commands - Style Guide

## Overview
This style guide defines the standard format and conventions for all Claude Code command documentation files in the `.claude/commands/` directory.

## Command Categories

### 1. Simple Commands
**Purpose**: Basic operations with straightforward execution steps
**Examples**: Session management commands (`session-start`, `session-current`, etc.)

**Template Structure:**
```markdown
# Claude Code Command: `/command-name`

## Purpose
[One-line description of what the command does]

## Execution Steps
1. [Step 1 description]
2. [Step 2 description]
3. [Step 3 description]

## [Optional Additional Sections]
- Guidelines
- Output Requirements
- Examples
```

### 2. Complex Commands
**Purpose**: Advanced operations requiring comprehensive documentation
**Examples**: Feature commands (`create-issues`, `work-on-task`, `auto-label-issues`)

**Template Structure:**
```markdown
# Claude Code Command: `/command-name`

## Command Overview
**Purpose**: [Detailed purpose statement]
**Syntax**: `/command-name <required> [--optional=value]`
**Examples**: 
- `/command-name example1` - Brief description
**Note**: Important usage information

## Command Description
[Detailed explanation of what the command does]

## System Instructions
### Role Definition
[Claude's role when executing this command]

### [Command Name] Workflow
[Detailed workflow phases and steps]

## Implementation Details
[Technical implementation information]

## Usage Examples
[Real-world usage scenarios]

## Configuration
[Configuration options and settings]

## Error Handling
[Common error scenarios and solutions]

## Best Practices
[Recommended usage patterns]
```

## Formatting Standards

### 1. Headers
- **Primary Header**: Always use `# Claude Code Command: `/command-name``
- **Section Headers**: Use `##` for main sections, `###` for subsections
- **Consistency**: All command files must start with the standardized header

### 2. Purpose Statement
- **Simple Commands**: One concise sentence describing the command's function
- **Complex Commands**: Detailed purpose statement in Command Overview section
- **Clarity**: Must be immediately understandable to new users

### 3. Syntax Documentation (Complex Commands Only)
- **Format**: `**Syntax**: /command-name <required> [--optional=value]`
- **Parameters**: Use `<>` for required, `[]` for optional
- **Flags**: Document all available flags with descriptions

### 4. Examples
- **Minimum**: At least one basic usage example
- **Format**: Use code blocks with brief descriptions
- **Variety**: Show different use cases and parameter combinations

### 5. Execution Steps
- **Numbering**: Use ordered lists (1., 2., 3.)
- **Clarity**: Each step should be actionable and specific
- **Nesting**: Use sub-bullets for detailed instructions
- **Error Handling**: Include conditional steps (if/else scenarios)

## Content Guidelines

### 1. Language and Tone
- **Clarity**: Use clear, direct language
- **Consistency**: Maintain consistent terminology across all commands
- **Professionalism**: Technical but accessible writing style

### 2. Code Examples
- **Formatting**: Use proper markdown code blocks with language specification
- **Comments**: Include explanatory comments where helpful
- **Completeness**: Examples should be functional and realistic

### 3. Cross-References
- **Command References**: Use backticks for command names (e.g., `/session-start`)
- **File References**: Use backticks for file paths and names
- **Linking**: Reference related commands when appropriate

## Documentation Depth Guidelines

### Simple Commands (Session Management)
**Required Sections:**
- Header
- Purpose
- Execution Steps

**Optional Sections:**
- Guidelines
- Output Requirements
- Examples

### Complex Commands (Feature Operations)
**Required Sections:**
- Header
- Command Overview
- Command Description
- System Instructions
- Implementation Details

**Optional Sections:**
- Usage Examples
- Configuration
- Error Handling
- Best Practices

## Quality Standards

### 1. Completeness
- All required sections must be present
- No placeholder text or "TODO" items
- Examples must be functional and tested

### 2. Accuracy
- All syntax examples must be correct
- Command behavior must be accurately described
- Error scenarios should be realistic

### 3. Maintainability
- Regular review and updates required
- Version information when applicable
- Clear ownership and responsibility

## File Organization

### 1. Naming Convention
- Use kebab-case for file names
- Match command name exactly (e.g., `session-start.md` for `/session-start`)
- Use `.md` extension for all command files

### 2. Directory Structure
```
.claude/commands/
├── session-start.md          # Simple command
├── session-current.md        # Simple command
├── session-end.md            # Simple command
├── create-issues.md          # Complex command
├── work-on-task.md           # Complex command
└── STYLE_GUIDE.md           # This document
```

### 3. Version Control
- All command files should be version controlled
- Use conventional commit messages for updates
- Tag major documentation changes

## Validation Checklist

### Pre-Publication Checklist
- [ ] Correct header format used
- [ ] Purpose statement is clear and concise
- [ ] All required sections are present
- [ ] Examples are functional and tested
- [ ] Syntax documentation is accurate
- [ ] Cross-references are valid
- [ ] Language is clear and professional
- [ ] File naming follows conventions

### Review Process
1. **Self-Review**: Author reviews against this style guide
2. **Peer Review**: Another team member reviews for clarity
3. **Testing**: Examples and syntax are tested
4. **Publication**: File is committed to repository

## Maintenance

### Regular Updates
- Review quarterly for accuracy
- Update examples as command behavior changes
- Ensure consistency with system changes

### Deprecation Process
1. Mark command as deprecated with clear notice
2. Provide migration path to replacement command
3. Remove after appropriate deprecation period

---

**Version**: 1.0  
**Last Updated**: 2025-06-17  
**Maintained By**: Development Team  
**Next Review**: 2025-09-17  

*This style guide ensures consistency, maintainability, and usability across all Claude Code command documentation.*