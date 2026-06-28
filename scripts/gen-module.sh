#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMPLATE_DIR="${ROOT_DIR}/scripts/templates/module"
MODULE="${1:-}"

usage() {
	echo "usage: $0 <module_name>"
	echo "example: make gen-module NAME=billing"
	exit 1
}

to_pascal() {
	local name="$1"
	local result=""
	local part
	IFS='_' read -ra parts <<< "$name"
	for part in "${parts[@]}"; do
		local first
		first="$(echo "${part:0:1}" | tr '[:lower:]' '[:upper:]')"
		result+="${first}${part:1}"
	done
	echo "$result"
}

render_template() {
	local tmpl="$1"
	local out="$2"
	sed \
		-e "s/{{MODULE}}/${MODULE}/g" \
		-e "s/{{PASCAL}}/${PASCAL}/g" \
		-e "s/{{MODULE_PATH}}/${MODULE}/g" \
		"$tmpl" >"$out"
}

patch_bootstrap() {
	local bootstrap="${ROOT_DIR}/internal/bootstrap/app.go"
	local import_path="github.com/radius/radius-backend/internal/${MODULE}"

	if ! grep -q "\"${import_path}\"" "$bootstrap"; then
		awk -v ipath="${import_path}" '
			/internal\/users/ && !done {
				print "\t\"" ipath "\""
				done = 1
			}
			{ print }
		' "$bootstrap" >"${bootstrap}.tmp" && mv "${bootstrap}.tmp" "$bootstrap"
	fi

	if ! grep -q "${MODULE}.NewModule()" "$bootstrap"; then
		awk -v mod="${MODULE}" '
			/storage\.NewModule\(\),/ && !done {
				print "\t\t" mod ".NewModule(),"
				done = 1
			}
			{ print }
		' "$bootstrap" >"${bootstrap}.tmp" && mv "${bootstrap}.tmp" "$bootstrap"
	fi
}

print_next_steps() {
	cat <<EOF

Module "${MODULE}" generated.

Test:
  make run
  curl -s http://localhost:8080/${MODULE}/hello

Next steps:
  1. Implement entity + repository in internal/${MODULE}/domain/
  2. Add ent/schema, then: make migrate-diff NAME=add_${MODULE}
  3. Implement internal/${MODULE}/infrastructure/db/postgres/
  4. Wire repositories in internal/${MODULE}/module.go wire()
  5. Replace or extend the hello endpoint in the controller

EOF
}

if [[ -z "$MODULE" ]]; then
	usage
fi

if [[ ! "$MODULE" =~ ^[a-z][a-z0-9_]*$ ]]; then
	echo "error: module name must match ^[a-z][a-z0-9_]*$ (got: ${MODULE})" >&2
	exit 1
fi

MODULE_DIR="${ROOT_DIR}/internal/${MODULE}"
if [[ -e "$MODULE_DIR" ]]; then
	echo "error: ${MODULE_DIR} already exists" >&2
	exit 1
fi

PASCAL="$(to_pascal "$MODULE")"

mkdir -p \
	"${MODULE_DIR}/domain" \
	"${MODULE_DIR}/application/dto" \
	"${MODULE_DIR}/application/services" \
	"${MODULE_DIR}/interface/api/rest" \
	"${MODULE_DIR}/infrastructure/db/postgres"

render_template "${TEMPLATE_DIR}/module.go.tmpl" "${MODULE_DIR}/module.go"
render_template "${TEMPLATE_DIR}/domain_errors.go.tmpl" "${MODULE_DIR}/domain/errors.go"
render_template "${TEMPLATE_DIR}/domain_entity.go.tmpl" "${MODULE_DIR}/domain/entity.go"
render_template "${TEMPLATE_DIR}/domain_repository.go.tmpl" "${MODULE_DIR}/domain/repository.go"
render_template "${TEMPLATE_DIR}/domain_transaction.go.tmpl" "${MODULE_DIR}/domain/transaction.go"
render_template "${TEMPLATE_DIR}/dto_hello.go.tmpl" "${MODULE_DIR}/application/dto/hello.go"
render_template "${TEMPLATE_DIR}/service.go.tmpl" "${MODULE_DIR}/application/services/${MODULE}_service.go"
render_template "${TEMPLATE_DIR}/controller.go.tmpl" "${MODULE_DIR}/interface/api/rest/${MODULE}_controller.go"
touch "${MODULE_DIR}/infrastructure/db/postgres/.gitkeep"

patch_bootstrap

(
	cd "$ROOT_DIR"
	go fmt "./internal/${MODULE}/..."
	go fmt "./internal/bootstrap/app.go" >/dev/null
)

echo "Generated internal/${MODULE}/"
print_next_steps
