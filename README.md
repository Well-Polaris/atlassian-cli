# Atlassian CLI

A comprehensive CLI tool for accessing Atlassian APIs, designed for use with Claude Code and automation workflows.

## Features

- **Jira** - Issues, projects, comments, worklogs, transitions (REST API)
- **Confluence** - Pages, spaces, comments, CQL search (REST API)
- **Jira Product Discovery (JPD)** - Ideas, insights (REST + GraphQL)
- **Goals** - Goals, metrics, status updates (GraphQL API)
- **Projects** - Atlas projects, goal linking (GraphQL API)
- **Compass** - Components, scorecards, raw GraphQL (GraphQL API)
- **Unified Search** - Search across Jira and Confluence (Rovo)

## Installation

```bash
# Clone and build
git clone https://github.com/peter/atlassian-cli.git
cd atlassian-cli
make build

# Or install to GOPATH/bin
make install
```

## Configuration

### Quick Start

```bash
# Initialize configuration
atlassian config init

# Edit .env with your credentials
```

### Environment Variables

Create a `.env` file (or set environment variables):

```bash
# Site URL (required)
ATLASSIAN_SITE_URL=https://yoursite.atlassian.net

# REST API auth (for Jira/Confluence)
ATLASSIAN_EMAIL=your-email@example.com
ATLASSIAN_API_TOKEN=your-api-token

# OAuth (for Goals/Projects GraphQL APIs)
ATLASSIAN_CLIENT_ID=your-client-id
ATLASSIAN_CLIENT_SECRET=your-client-secret
ATLASSIAN_ACCESS_TOKEN=your-access-token
ATLASSIAN_REFRESH_TOKEN=your-refresh-token
```

### Getting Credentials

**REST API Token:**
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Create API token
3. Set `ATLASSIAN_EMAIL` and `ATLASSIAN_API_TOKEN`

**OAuth (for GraphQL APIs):**
1. Go to https://developer.atlassian.com/console/myapps/
2. Create an OAuth 2.0 app
3. Add scopes: `read:me`, `read:jira-work`, `read:confluence-content.all`, `offline_access`
4. Run `atlassian auth login`

## Usage

### Jira

```bash
# Issues
atlassian jira issue get PROJ-123
atlassian jira issue create -p PROJ -t Task -s "Summary"
atlassian jira issue update PROJ-123 -s "New summary"
atlassian jira issue search "project = PROJ AND status = Open"
atlassian jira issue transition PROJ-123 "In Progress"

# Comments
atlassian jira comment list PROJ-123
atlassian jira comment add PROJ-123 "My comment"

# Projects
atlassian jira project list
atlassian jira project get PROJ

# Worklog
atlassian jira worklog PROJ-123 2h
```

### Confluence

```bash
# Pages
atlassian confluence page get 123456789
atlassian confluence page list --space 12345
atlassian confluence page create --space 12345 --title "My Page" --body "<p>Content</p>"
atlassian confluence page update 123456789 --body "<p>Updated</p>"

# Spaces
atlassian confluence space list
atlassian confluence space get 12345

# Comments
atlassian confluence comment list 123456789
atlassian confluence comment add 123456789 "My comment"

# Search (CQL)
atlassian confluence search "type = page AND title ~ 'meeting'"
```

### Goals (GraphQL)

```bash
# List and get
atlassian goals list
atlassian goals get goal-id

# Create and update
atlassian goals create --name "Q1 Revenue Target" --due 2024-03-31
atlassian goals update goal-id --state on_track

# Metrics
atlassian goals metric update metric-id 75.5

# Status updates
atlassian goals status list goal-id
atlassian goals status add goal-id --message "On track for Q1" --state on_track
```

### Projects (GraphQL)

```bash
# List and get
atlassian projects list
atlassian projects get project-id

# Create and update
atlassian projects create --name "Website Redesign"
atlassian projects update project-id --state in_progress

# Link goals
atlassian projects link-goal project-id goal-id
atlassian projects unlink-goal project-id goal-id

# Status
atlassian projects status project-id --message "Phase 1 complete"
```

### Compass (GraphQL)

Compass is GraphQL-only and uses OAuth. Run `atlassian auth login` first.

```bash
# Components
atlassian compass component list
atlassian compass component list --query "polaris" --limit 20
atlassian compass component get "ari:cloud:compass:<cloudId>:component/..."

# Scorecards
atlassian compass scorecard list

# Raw GraphQL — the escape hatch for metrics, teams, dependencies and
# all mutations. The gateway requires a NAMED operation.
atlassian compass query 'query Q($c:String!){ compass { searchComponents(cloudId:$c, query:{first:5}){ ... on CompassSearchComponentConnection { totalCount } } } }' --vars '{"c":"<cloudId>"}'
atlassian compass query --file ./mutation.graphql --vars '{"input":{...}}'
```

### Jira Product Discovery

```bash
# Projects — list every JPD project (the access the Atlassian MCP omits)
atlassian jpd projects list
atlassian jpd projects list --json

# Ideas (uses Jira REST API)
atlassian jpd ideas list -p PROJ
atlassian jpd ideas get PROJ-123
atlassian jpd ideas create -p PROJ -s "New feature idea"

# Insights (uses GraphQL)
atlassian jpd insights list PROJ-123

# Comments
atlassian jpd comment PROJ-123 "Great idea!"
```

### Issue Links (Jira ↔ JPD)

Link regular Jira tickets to Product Discovery ideas and vice versa — something
the Atlassian MCP cannot do. JPD ideas are Jira issues, so any key works on
either side.

```bash
# List the link types available in your instance
atlassian link types

# Link a Jira ticket to a JPD idea. Reads "<from> <relation> <to>" —
# for type "Blocks", PROJ-123 blocks IDEA-45.
atlassian link create PROJ-123 IDEA-45 --type "Blocks"
atlassian link create PROJ-123 IDEA-45 --type "Relates" --comment "delivers this idea"

# List links on an issue or idea
atlassian link list IDEA-45

# Delete a link by its ID (shown in 'link list')
atlassian link delete 10042
```

Projects can also be filtered by type directly via Jira:

```bash
atlassian jira project list --type product_discovery
```

### Unified Search

```bash
# Search across Jira and Confluence
atlassian search "quarterly report"

# Fetch by ARI (Atlassian Resource Identifier)
atlassian fetch "ari:cloud:jira:cloudid:issue/10107"
```

### Configuration

```bash
# Show current config
atlassian config show

# Initialize config file
atlassian config init
atlassian config init --global  # ~/.atlassian-cli/.env

# Show config paths
atlassian config path
```

### Authentication

```bash
# OAuth login flow
atlassian auth login

# Refresh access token
atlassian auth refresh

# Check auth status
atlassian auth status
```

## Output Formats

Most commands support `--json` flag for JSON output:

```bash
atlassian jira issue get PROJ-123 --json | jq '.fields.summary'
```

## Version

```bash
atlassian --version
# atlassian version 0.1.0
```

## License

MIT
