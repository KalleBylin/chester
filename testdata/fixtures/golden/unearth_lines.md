# db/queries.go:112-115

- L112-L113
  commit: aaaaaaa
  source: PR #77 Bypass ORM for hot path query
  summary: Use raw SQL for this lookup because the ORM path regressed under PostgreSQL 13.
  summary-source: pr body

- L114-L115
  commit: bbbbbbb
  source: direct
  summary: Normalize placeholders so the query stays portable across drivers.
  summary-source: commit body

## File Notes
- PR #77 Bypass ORM for hot path query
  @maintainer: Bypassing ORM due to Postgres 13 bug.
  @reviewer: Keep until PG13 is no longer supported.
