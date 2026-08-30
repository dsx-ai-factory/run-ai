#!/bin/sh
set -e
S=/opt/keycloak/conf/secrets
CONF=/conf/keycloak.conf
: > "$CONF"
for file in "$S"/*; do
  [ -f "$file" ] || continue
  name=$(basename "$file")
  case "$name" in
    KEYCLOAK_ADMIN | KEYCLOAK_ADMIN_PASSWORD) continue ;;
    KC_*) key=$(printf '%s' "$name" | sed 's/^KC_//' | tr '[:upper:]_' '[:lower:]-') ;;
    *)    key="$name" ;;
  esac
  printf '%s=%s\n' "$key" "$(cat "$file")" >> "$CONF"
done
