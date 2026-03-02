# internal/auth/session.go

- 1111111,2222222 PR #98 Split session store
  why: Move persistence behind a store interface so auth and admin flows stop sharing transaction logic.

- 3333333 direct
  why: Fix nil panic when malformed cookie omits signature segment.
