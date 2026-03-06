#123 [pr] Harden session invalidation
https://github.com/acme/chester/pull/123

## Body
Force immediate revocation after password reset.

## Comments
- @alice
  Session cookies survive password reset on main.

- @carol
  Confirmed this also affects API tokens.

## Reviews
- APPROVED @bob
  Keep this branch-specific invalidation until legacy clients are removed.

## Review Comments
- internal/auth/session.go @bob (L88)
  Keep the branch-specific invalidation until legacy clients are removed.
