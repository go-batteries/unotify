#!/bin/bash
#
# Merge all swagger json files into one

OUTPUT_FILE="openapiv2/apidocs.json"

function merge() {
  swagger mixin \
    $(find ./openapiv2 -type f -name '*.swagger.json' | tr '\n' ' ') 1>&2 >/dev/null > "$OUTPUT_FILE"
}

function serve() {
  docker run --rm -p 8080:8080 \
  	-e SWAGGER_JSON="/${OUTPUT_FILE}" \
  	-v ./openapiv2/:/openapiv2 \
  	swaggerapi/swagger-ui
}

if [[ "$1" != "skip-merge" ]]; then
  echo "merging swagger.json"
  merge
fi

if [[ -s "$OUTPUT_FILE" ]]; then
  echo "preparing to serve"
  serve
fi
