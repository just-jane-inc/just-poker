rm -rf openapi_client
source .venv/bin/activate
openapi-generator-cli generate --additional-properties=library=httpx -i 'http://localhost:7653/swagger/openapi.json' -g python -o async-lib
mv async-lib/openapi_client .
