#!/usr/bin/env bash
set -euo pipefail

# Can be http://localhost:7653/swagger/openapi.json if running server
SPEC="${1:-../server/docs/swagger.json}"
# Can be say, npx openapi-generator-cli,
GENERATOR="${OPENAPI_GENERATOR:-openapi-generator-cli}"
OUT="generated"

rm -rf "generated"
"${GENERATOR}" generate \
  --additional-properties=packageName=JustPoker.OpenApi,library=generichost,targetFramework=net10.0,nullableReferenceTypes=true,useDateTimeOffset=true \
  -i "${SPEC}" \
  -g csharp \
  --output "generated"
