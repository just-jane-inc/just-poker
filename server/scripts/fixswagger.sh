#!/usr/bin/env bash
set -euo pipefail

# YAML
awk '
/^[[:space:]]*oneOf:[[:space:]]*$/ {
    in_oneof = 1
    item_indent = -1
    oneof_indent = match($0, /[^ ]/) - 1
    print
    next
}
in_oneof && $0 ~ /^[[:space:]]*$/ { print; next }
in_oneof {
    indent = match($0, /[^ ]/) - 1
    is_item = ($0 ~ /^[[:space:]]*-[[:space:]]/)

    if (indent < oneof_indent || (indent == oneof_indent && !is_item)) {
        in_oneof = 0
    } else {
        if (is_item && item_indent == -1) item_indent = indent

        if (is_item && indent == item_indent &&
            $0 ~ /^[[:space:]]*-[[:space:]]*type:[[:space:]]*object[[:space:]]*$/) {
            next
        }
    }
}
{ print }
' docs/swagger.yaml >docs/swagger.yaml.tmp && mv docs/swagger.yaml.tmp docs/swagger.yaml

# JSON
awk '
function drop_trailing_comma() {
    if (n > 0) sub(/,[[:space:]]*$/, "", out[n])
}
{
    if ($0 ~ /"oneOf"[[:space:]]*:[[:space:]]*\[/) {
        in_oneof = 1
        oneof_indent = match($0, /[^ ]/) - 1
        out[++n] = $0
        next
    }

    if (in_oneof) {
        indent = match($0, /[^ ]/) - 1

        if (indent <= oneof_indent && $0 ~ /^[[:space:]]*\]/) {
            in_oneof = 0
            out[++n] = $0
            next
        }

        if ($0 ~ /^[[:space:]]*\{[[:space:]]*$/) {
            open = $0
            if ((getline second) <= 0) { out[++n] = open; next }

            if (second !~ /^[[:space:]]*"type"[[:space:]]*:[[:space:]]*"object"[[:space:]]*$/) {
                out[++n] = open; out[++n] = second; next
            }

            if ((getline close_line) <= 0) { out[++n] = open; out[++n] = second; next }

            if (close_line !~ /^[[:space:]]*\}[[:space:]]*,?[[:space:]]*$/) {
                out[++n] = open; out[++n] = second; out[++n] = close_line; next
            }

            if (close_line !~ /,[[:space:]]*$/) drop_trailing_comma()
            next
        }
    }
    out[++n] = $0
}
END { for (i = 1; i <= n; i++) print out[i] }
' docs/swagger.json >docs/swagger.json.tmp && mv docs/swagger.json.tmp docs/swagger.json
