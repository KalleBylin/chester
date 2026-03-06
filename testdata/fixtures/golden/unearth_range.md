# main..feature

- PR #98 Split session store
  first-commit: 1111111
  summary: Move persistence behind a store interface so auth and admin flows stop sharing transaction logic.
  summary-source: pr body

- PR #151 Invalidate sessions on password reset
  first-commit: 3333333
  summary: Force immediate credential revocation across web and API flows.
  summary-source: pr body

- direct
  commit: 4444444
  summary: Regenerate golden CLI output for the auth fixtures.
  summary-source: commit body
