# GoSight UI Roadmap

This document outlines the UI development roadmap for GoSight. It is structured by feature area, with prioritized foundational work listed first.

## Foundational UI

### Alerts
- [ ] Alert Rule Builder
  - [ ] Build the alert rule form (expression, scope, level, actions)
  - [ ] Wire up to API (create, update)
  - [ ] Validate required fields
  - [ ] Add success/failure feedback

- [ ] Alert History Page
  - [ ] Create table with columns: Rule, Level, Status, FiredAt, ResolvedAt
  - [ ] Add filters: time range, rule name, endpoint, status
  - [ ] Paginate results
  - [ ] Click-to-expand for detail view

### Log Explorer
- [ ] Build table: Timestamp, Level, Message, Source, Endpoint
- [ ] Add filter inputs: date range, endpoint, level, message content
- [ ] Pagination and search
- [ ] Export to CSV and JSON
- [ ] Backend: `/api/v1/logs` with filters

### Admin
- [ ] IAM Center
  - [ ] Users page: email, roles, status, MFA
  - [ ] Roles page: name, description, permissions
  - [ ] Permissions page: ID, description
  - [ ] Assign roles to users
  - [ ] Attach permissions to roles

- [ ] Audit Log Viewer
  - [ ] Table with: User, Action, Resource, Timestamp
  - [ ] Filters by user, action, time range

### General UI Cleanup
- [ ] Unify color usage with Tailwind
- [ ] Fix dark mode toggle and contrast
- [ ] Extract reusable components (cards, tables, etc.)
- [ ] Clean up unused CSS/JS

---

## Advanced UX

### Incidents View
- [ ] Create `/incidents` page
- [ ] Show active alerts requiring review/remediation
- [ ] Detail view with:
  - [ ] Triggering rule and context
  - [ ] Timeline (when it started)
  - [ ] Related logs, metrics, events
  - [ ] Optional remediation status

### Endpoint Container Details
- [ ] Build `/endpoints/{ctr-id}` view
- [ ] Render metrics and logs like a host detail page
- [ ] Include uptime, CPU, memory, status, runtime

### Container Actions Menu
- [ ] Dropdown for Podman/Docker commands
- [ ] Actions: restart, stop, inspect, stats
- [ ] Send to backend securely (command allowlist)

---

## Global Site Search

- [ ] Implement autosuggest dropdown with categories
  - [ ] Agents, Containers, Metrics, Logs
- [ ] Keyboard navigation for dropdown
- [ ] Search results page with grouped results
- [ ] Optional: fuzzy search or backend query

---

## Reports Page

- [ ] Build `/reports` interface
- [ ] Time range selector
- [ ] Sectioned summaries:
  - [ ] CPU, Memory, Disk, Network
  - [ ] Logs and event stats
- [ ] Export to PDF or printable HTML
- [ ] Branding support (whitelabel logo/header)

---

## Admin Configuration and Whitelabeling

- [ ] Global Config Editor
  - [ ] Load current server config (YAML or JSON)
  - [ ] Editable form or code editor
  - [ ] Submit config changes via API

- [ ] Whitelabeling
  - [ ] Upload logo, favicon
  - [ ] Custom app name and footer text
  - [ ] Color theme overrides

---

## Notes

- All UI must remain fully responsive and dark mode compatible.
- Tailwind CSS and Flowbite are used for styling and components.
- Prefer modular HTML templates and minimal JS for maintainability.