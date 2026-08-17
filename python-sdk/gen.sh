rm -rf src/openapi_client/
source .venv/bin/activate
openapi-generator-cli generate \
  --additional-properties=library=httpx,packageName=openapi_client,generateSourceCodeOnly=true \
  -i 'http://localhost:7256/swagger/openapi.json' \
  -g python \
  --output src
#mv async-lib/openapi_client src/poker_bot
