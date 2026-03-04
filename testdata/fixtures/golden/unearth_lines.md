# db/queries.go:112-115

- L112-L113
  commit: aaaaaaa
  source: PR #77 Bypass ORM for hot path query
  why: Use raw SQL for this lookup because the ORM path regressed under PostgreSQL 13.
  notes: @maintainer: Bypassing ORM due to Postgres 13 bug. | @reviewer: Keep until PG13 is no longer supported.

- L114-L115
  commit: bbbbbbb
  source: direct
  why: Normalize placeholders so the query stays portable across drivers.
