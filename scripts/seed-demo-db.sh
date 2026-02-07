#!/usr/bin/env bash
# Creates a demo database with realistic dummy data for screenshots.
# Usage: ./scripts/seed-demo-db.sh [db-path]
#
# The database path defaults to ./demo.db

set -euo pipefail

DB="${1:-./demo.db}"
BUJO="./bujo"

if [ ! -f "$BUJO" ]; then
  echo "Building bujo binary..."
  go build -o "$BUJO" ./cmd/bujo
fi

rm -f "$DB"
echo "Creating demo database at $DB"

b() { "$BUJO" --db-path "$DB" "$@"; }

# --- Day Context (today and recent days) ---
today=$(date +%Y-%m-%d)
yesterday=$(date -v-1d +%Y-%m-%d)
two_days_ago=$(date -v-2d +%Y-%m-%d)
three_days_ago=$(date -v-3d +%Y-%m-%d)
four_days_ago=$(date -v-4d +%Y-%m-%d)
five_days_ago=$(date -v-5d +%Y-%m-%d)
six_days_ago=$(date -v-6d +%Y-%m-%d)

b mood set --date "$today" "Focused"
b mood set --date "$yesterday" "Energised"
b mood set --date "$two_days_ago" "Creative"
b mood set --date "$three_days_ago" "Calm"
b mood set --date "$four_days_ago" "Productive"

# --- Today's entries ---
b add --date "$today" \
  ". Review pull request for auth module" \
  ". !!! Deploy staging environment" \
  "- Team standup: discussed migration timeline" \
  "o Lunch with Sarah at 12:30" \
  ". !! Update API documentation" \
  ". Write unit tests for payment service" \
  "- Consider switching to structured logging" \
  "? Should we use Redis or Memcached for session store" \
  ". Fix flaky CI test in user-service" \
  "o Design review at 3pm"

# Mark some as done
b done 1   # Review PR
b done 6   # Write unit tests

# Cancel one
b cancel 10  # Design review cancelled

# --- Yesterday's entries ---
b add --date "$yesterday" \
  ". Set up monitoring dashboards" \
  ". !! Refactor database connection pooling" \
  "- Pair programming session with Alex went well" \
  ". Sprint retrospective" \
  ". Investigate memory leak in worker process" \
  ". Update dependencies to latest versions" \
  "- New team member joining next week — prepare onboarding docs"

b done 11  # monitoring dashboards
b done 14  # sprint retro
b done 16  # update deps

# --- 2 days ago ---
b add --date "$two_days_ago" \
  ". !!! Fix production error in checkout flow" \
  ". Draft technical design for notifications" \
  "- Good progress on the notification system design" \
  ". Architecture review meeting" \
  ". Add rate limiting to public API"

b done 18  # fix production error
b done 21  # architecture review

# --- 3 days ago ---
b add --date "$three_days_ago" \
  ". Create database migration scripts" \
  "- Learned about Go's new range-over-func feature" \
  ". !! Set up integration test environment" \
  "o Team lunch at the new ramen place"

b done 23  # migration scripts

# --- 4 days ago ---
b add --date "$four_days_ago" \
  ". Implement user profile API endpoint" \
  ". Write acceptance tests for signup flow" \
  "- Performance benchmarks show 3x improvement after indexing" \
  "o Product planning session"

b done 27  # user profile API
b done 28  # acceptance tests

# --- 5 days ago ---
b add --date "$five_days_ago" \
  ". Configure CI/CD pipeline for new service" \
  "- Kubernetes cluster upgrade went smoothly" \
  ". Review and merge open PRs"

b done 31  # CI/CD pipeline
b done 33  # merge PRs

# --- 6 days ago ---
b add --date "$six_days_ago" \
  ". Set up project scaffolding for notification service" \
  "- Good discussion about event-driven architecture" \
  "o Coffee chat with engineering manager"

b done 34  # project scaffolding

# --- Overdue entries (from earlier) ---
seven_days_ago=$(date -v-7d +%Y-%m-%d)
ten_days_ago=$(date -v-10d +%Y-%m-%d)

b add --date "$seven_days_ago" \
  ". !! Write blog post about Go testing patterns" \
  ". Update team wiki with deployment runbook"

b add --date "$ten_days_ago" \
  ". !!! Security audit for authentication service" \
  ". Clean up unused feature flags"

# --- Questions ---
twelve_days_ago=$(date -v-12d +%Y-%m-%d)
b add --date "$twelve_days_ago" \
  "? What monitoring tool should we adopt for the new platform" \
  "? How do we handle backwards compatibility for API v2"

# Answer one
b answer 42 "We'll use Grafana with Prometheus for metrics and Loki for logs"

# --- Habits ---
# Habits are created implicitly via --yes on first log

# Log habits over the past 14 days for realistic calendar
for i in $(seq 0 13); do
  d=$(date -v-${i}d +%Y-%m-%d)

  # Exercise: ~5 days/week
  if [ $((i % 7)) -ne 0 ] && [ $((i % 7)) -ne 3 ]; then
    b habit log "Exercise" --yes --date "$d"
  fi

  # Reading: almost every day
  if [ $((i % 8)) -ne 0 ]; then
    b habit log "Reading" --yes --date "$d"
  fi

  # Meditation: ~4 days/week
  if [ $((i % 3)) -ne 0 ]; then
    b habit log "Meditation" --yes --date "$d"
  fi

  # Journaling: every day
  b habit log "Journaling" --yes --date "$d"

  # Water: variable count per day (3-8)
  count=$(( (i * 7 + 3) % 6 + 3 ))
  for j in $(seq 1 $count); do
    b habit log "Water intake" --yes --date "$d"
  done
done

b habit set-goal "Water intake" 8

# --- Lists ---
b list create "Reading List"
b list add "Reading List" "Designing Data-Intensive Applications"
b list add "Reading List" "The Pragmatic Programmer"
b list add "Reading List" "Clean Architecture"
b list add "Reading List" "Domain-Driven Design"
b list add "Reading List" "Release It!"

# Mark some reading list items done
b list done 1
b list done 2

b list create "Home Projects"
b list add "Home Projects" "Fix kitchen faucet"
b list add "Home Projects" "Organise garage"
b list add "Home Projects" "Paint bedroom"
b list add "Home Projects" "Install shelf in office"

b list done 6

b list create "Side Project Ideas"
b list add "Side Project Ideas" "CLI tool for time tracking"
b list add "Side Project Ideas" "Personal finance dashboard"
b list add "Side Project Ideas" "Recipe API with Golang"

# --- Goals ---
current_month=$(date +%Y-%m)

b goal add --month "$current_month" "Ship v2.0 of the notification service"
b goal add --month "$current_month" "Complete Go certification course"
b goal add --month "$current_month" "Read 3 technical books"
b goal add --month "$current_month" "Mentor two junior developers"
b goal add --month "$current_month" "Run a half marathon"

# Complete one goal
b goal done 1

echo ""
echo "Demo database created at $DB"
echo "Run with: ./bujo --db-path $DB tui"
echo "Or desktop: wails dev -- --db-path $DB"
