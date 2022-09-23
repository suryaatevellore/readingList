#!/bin/bash
set -e # Exit with nonzero exit code if anything fails
set -o pipefail
set -o errexit

dockerSQLUtil="sqlite-utils"

TZ=UTC

buildDocker() {
  docker build --tag "$dockerSQLUtil" --file sqlite-utils.Dockerfile .
}

sql-utils() {
  docker run \
    -i \
    -u"$(id -u):$(id -g)" \
    -v"$(pwd):/wd" \
    -w /wd \
    "$dockerSQLUtil" \
    "$@"
}


makeDB() {
  local db="$1"
  rm -rf "$db" || true
  sql-utils insert "$db" read readingList.csv --csv
  sql-utils optimize "$db"
}

commitDB() {
  local dbBranch="db"
  local db="$1"
  local tempDB="$(mktemp)"
  git config user.name "Automated"
  git config user.email "actions@users.noreply.github.com"
  git branch -D "$dbBranch" || true
  git checkout --orphan "$dbBranch"
  mv "$db" "$tempDB"
  rm -rf *
  mv "$tempDB" "$db"
  git add -f "$db"
  git commit "$db" -m "push db"
  git push origin "$dbBranch" -f
}


publishDB() {
  local dockerDatasette="datasette"
  docker build --tag "$dockerDatasette" --pull --file datasette.Dockerfile .
  docker run \
    -v"$(pwd):/wd" \
    -w /wd \
    "$dockerDatasette" \
    publish vercel "$db" --token "$VERCEL_TOKEN" --project=reading-list --install=datasette-vega
}

run() {
  local db="readingList.db"
  makeDB "$db"
  publishDB "$db"
  commitDB "$db"

}

buildDocker
run "$@"
