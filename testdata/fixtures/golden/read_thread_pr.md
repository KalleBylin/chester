#123 [pr] Harden session invalidation
https://github.com/acme/chester/pull/123

## Body
Force immediate revocation after password reset.

## Thread
- @alice
  Session cookies survive password reset on main.

- review/APPROVED @bob
  Keep this branch-specific invalidation until legacy clients are removed.

- @carol
  Confirmed this also affects API tokens.
