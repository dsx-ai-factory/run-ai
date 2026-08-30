#!/bin/bash

COVER_FILE="cover.out"

calc_badge_color() {
  percentage=$(echo "${1}" | awk -F '%' '{print $1}' | awk -F '.' '{print $1}')
  int=$((percentage))
  if [ $int -ge 90 ]; then
    echo "brightgreen"
  elif [ $int -ge 80 ]; then
    echo "green"
  elif [ $int -ge 70 ]; then
    echo "yellowgreen"
  elif [ $int -ge 60 ]; then
    echo "yellow"
  elif [ $int -ge 50 ]; then
    echo "orange"
  else
    echo "red"
  fi
}

filter_cover() {
  grep_str=$1
  filtered_file="filtered_${2}.out"
  echo "$filtered_file"
  echo "mode: set" > "$filtered_file"
  grep "$grep_str" $COVER_FILE >> "$filtered_file"
}

create_cover_output() {
  filtered_file="filtered_${1}.out"
  coverage_file="coverage_${1}.out"
  echo "filtered file: $filtered_file coverage file: $coverage_file"
  go tool cover -func="$filtered_file" -o="$coverage_file" > /dev/null
  if [ $? -ne 0 ]; then
    echo "Failed to create coverage file"
    rm "$filtered_file"
    exit 1
  fi
  rm "$filtered_file"
}

get_coverage_url() {
  name=$1
  create_cover_output "$name"
  coverage_file="coverage_${name}.out"
  cov=$(grep total "$coverage_file" | awk '{print $3}')
  badge_color=$(calc_badge_color "$cov")
  echo "badge_color: $badge_color cov: $cov"
  badge_url="${name}-${cov}25-blue?color=${badge_color}"
  echo "url: $badge_url"
  base_url="${name}-[0-9]\{1,\}\.[0-9]\{1,\}%25-blue?color=[A-Za-z]\{1,\}"
  echo "base_url: $base_url"
  sed -i '' "s/${base_url}/${badge_url}/" README.md
  rm "$coverage_file"
}

go test -v ./... -coverprofile cover.out

grep -v "internal/autogenerate" $COVER_FILE > filtered_total.out
filter_cover "runai-cli/cmd/" cmd
filter_cover "runai-cli/internal/" internal
filter_cover "runai-cli/pkg/" pkg

get_coverage_url total
get_coverage_url cmd
get_coverage_url internal
get_coverage_url pkg

rm cover.out
