rm -rf openapi_client
source .venv/bin/activate
openapi-generator-cli generate -i 'http://localhost:7653/swagger/openapi.json' -g python -o lib
mv lib/openapi_client .
