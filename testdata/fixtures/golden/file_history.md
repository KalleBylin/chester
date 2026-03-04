# internal/auth/session.go

- PR #98 Split session store
  commits: 1111111,2222222
  why: Move persistence behind a store interface so auth and admin flows stop sharing transaction logic.

- direct
  commit: 3333333
  why: Fix nil panic when malformed cookie omits signature segment.
