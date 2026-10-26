#!/bin/sh
set -eu

# Relais de compatibilité. L'installateur maintenu reste dans scripts/.
# Avec une copie du dépôt, utiliser le script de cette même version.
if [ -z "${NODALIBRE_INSTALL_SCRIPT_URL:-}" ] && [ -f "$0" ]; then
  SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
  if [ -f "$SCRIPT_DIR/scripts/install.sh" ]; then
    exec sh "$SCRIPT_DIR/scripts/install.sh" "$@"
  fi
fi

SCRIPT_URL="${NODALIBRE_INSTALL_SCRIPT_URL:-https://raw.githubusercontent.com/AnARCHIS12/Libre-claude/main/scripts/install.sh}"
TMP_FILE="$(mktemp "${TMPDIR:-/tmp}/nodalibre-install.XXXXXX")"

cleanup() {
  rm -f "$TMP_FILE"
}
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$SCRIPT_URL" -o "$TMP_FILE"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP_FILE" "$SCRIPT_URL"
else
  printf '%s\n' "Erreur: curl ou wget est requis pour charger l'installateur Nodalibre." >&2
  exit 1
fi

sh "$TMP_FILE" "$@"
