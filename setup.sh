#!/bin/bash
#
# setup project

set -e

function setup_base() {
  touch README.md
  touch Makefile

  mkdir -p app/{core,web}
  mkdir -p cmd/{server,worker,cli}


  mkdir -p protos/{core,web}
  mkdir -p openapiv2
}


function setup_required_protos() {
  echo "building directories for google api protos"
  
  local google_api_dir_root="protos/includes/googleapis"
  local grpc_ecosystem_dir_root="protos/includes/grpc_ecosystem/protoc-gen-openapiv2"

  mkdir -p "${google_api_dir_root}/google/"{api,protobuf}
  mkdir -p "${grpc_ecosystem_dir_root}/options"

  ls "${google_api_dir_root}/google/api"
  ls "${grpc_ecosystem_dir_root}/options"

	curl "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto" \
		-o "${google_api_dir_root}/google/api/http.proto" \

	curl "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto" \
		-o "${google_api_dir_root}/google/api/annotations.proto"

	curl "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/descriptor.proto" \
		-o "${google_api_dir_root}/google/protobuf/descriptor.proto"

	curl "https://raw.githubusercontent.com/protocolbuffers/protobuf/main/src/google/protobuf/struct.proto" \
		-o "${google_api_dir_root}/google/protobuf/struct.proto"

	curl "https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto" \
	  -o "${grpc_ecosystem_dir_root}/options/annotations.proto"

	curl "https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto" \
		-o "${grpc_ecosystem_dir_root}/options/openapiv2.proto"
}

function setup_go_protogen_execs() {
  export GOBIN=$(go env GOBIN) 

  go mod download google.golang.org/protobuf
  go install google.golang.org/protobuf/cmd/protoc-gen-go
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
}

setup_base
setup_required_protos
setup_go_protogen_execs

echo "run go mod tidy"
