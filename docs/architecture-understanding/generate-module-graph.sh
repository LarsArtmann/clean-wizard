#!/usr/bin/env bash
# Generates the internal module-dependency graph (Mermaid) from the Go build
# graph. Regenerate after structural changes:
#
#   bash docs/architecture-understanding/generate-module-graph.sh
#
set -euo pipefail

module_path=$(go list -m)

echo "flowchart TD"

while IFS= read -r pkg; do
	short="${pkg#"$module_path"/}"
	[ "$short" = "$pkg" ] && short="$pkg"
	id=$(printf '%s' "$short" | tr '/.-' '___')

	deps=$(go list -f '{{range .Imports}}{{println .}}{{end}}' "$pkg" | grep "$module_path/internal" || true)
	while IFS= read -r dep; do
		[ -z "$dep" ] && continue
		dshort="${dep#"$module_path"/}"
		did=$(printf '%s' "$dshort" | tr '/.-' '___')
		printf '    %s["%s"] --> %s["%s"]\n' "$id" "$short" "$did" "$dshort"
	done <<<"$deps"
done < <(go list ./internal/... ./cmd/... | sort)
