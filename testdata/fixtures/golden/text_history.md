# text "SessionStore" in internal/auth/session.go

- PR #98 Split session store
  commits: 1111111
  summary: Move persistence behind a store interface so auth and admin flows stop sharing transaction logic.
  summary-source: pr body

- direct
  commit: 2222222
  summary: Rename SessionStore helper to match new package layout.
  summary-source: commit body
