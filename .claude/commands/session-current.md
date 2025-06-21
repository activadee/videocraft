# Claude Code Command: `/session-current`

## Purpose
Display current development session status and progress information

## Execution Steps
1. Check if `.claude/sessions/.current-session` exists
2. If no active session:
   - Inform user no session is active
   - Suggest starting one with `/session-start`
3. If active session exists, display:
   - Session name and filename
   - Duration since start (calculated)
   - Last few updates
   - Current goals/tasks
   - Available session commands

## Output Requirements
- Keep information concise and well-organized
- Present status in easy-to-read format