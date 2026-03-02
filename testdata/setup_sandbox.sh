#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <repo-dir>" >&2
  exit 1
fi

repo_dir="$1"

mkdir -p "$repo_dir"
cd "$repo_dir"

git init -b main >/dev/null
git config user.name "Chester Test"
git config user.email "chester@example.com"

mkdir -p internal/auth db docs

cat <<'EOF' > internal/auth/session.go
package auth

func ValidateSession(token string) bool {
	return token != ""
}
EOF

cat <<'EOF' > db/queries.go
package db

func LookupUserQuery() string {
	return "SELECT id FROM users WHERE email = $1"
}
EOF

cat <<'EOF' > docs/notes.md
baseline
EOF

git add internal/auth/session.go db/queries.go docs/notes.md
git commit -m "Initial repository layout" >/dev/null

echo "direct main note" >> docs/notes.md
git add docs/notes.md
git commit -m "Add direct baseline note" >/dev/null

range_from="$(git rev-parse HEAD)"

git checkout -b feature/raw-sql >/dev/null
cat <<'EOF' > db/queries.go
package db

func LookupUserQuery() string {
	query := "SELECT id FROM users WHERE email = $1"
	return query
}

func LookupSessionQuery() string {
	return "SELECT session_id FROM sessions WHERE token = $1"
}
EOF
git add db/queries.go
git commit -m "Prepare raw SQL path" >/dev/null

cat <<'EOF' > db/queries.go
package db

func LookupUserQuery() string {
	query := "SELECT id FROM users WHERE email = $1"
	return query
}

func LookupSessionQuery() string {
	return "SELECT session_id FROM sessions WHERE token = $1"
}

func HotPathQuery() string {
	return "SELECT id FROM sessions WHERE token = $1"
}
EOF
git add db/queries.go
git commit -m "Add hot path query helper" >/dev/null
git checkout main >/dev/null
git merge --squash feature/raw-sql >/dev/null
git commit -m "Bypass ORM for hot path query (#77)" >/dev/null
squash_sha="$(git rev-parse HEAD)"
git branch -D feature/raw-sql >/dev/null

git checkout -b feature/session-store >/dev/null
cat <<'EOF' > internal/auth/session.go
package auth

type SessionStore interface {
	Invalidate(token string) error
}

func ValidateSession(token string) bool {
	return token != ""
}
EOF
git add internal/auth/session.go
git commit -m "Introduce session store interface" >/dev/null

cat <<'EOF' > internal/auth/session.go
package auth

type SessionStore interface {
	Invalidate(token string) error
}

func ValidateSession(token string) bool {
	return token != ""
}

func InvalidateSession(store SessionStore, token string) error {
	return store.Invalidate(token)
}
EOF
git add internal/auth/session.go
git commit -m "Add invalidate helper" >/dev/null
git checkout main >/dev/null
git merge --no-ff feature/session-store -m "Merge pull request #98 from feature/session-store" >/dev/null
merge_sha="$(git rev-parse HEAD)"
git branch -D feature/session-store >/dev/null

echo "golden output refresh" >> docs/notes.md
git add docs/notes.md
git commit -m "Regenerate golden CLI output" >/dev/null
direct_sha="$(git rev-parse HEAD)"

cat <<EOF > .chester-sandbox-manifest
SESSION_FILE=internal/auth/session.go
DB_FILE=db/queries.go
RANGE_FROM=$range_from
RANGE_TO=HEAD
BLAME_START=3
BLAME_END=8
SQUASH_SHA=$squash_sha
MERGE_SHA=$merge_sha
DIRECT_SHA=$direct_sha
EOF
