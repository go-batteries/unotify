pre.install:	
	go get google.golang.org/protobuf 
	go get google.golang.org/genproto
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway


gen.models.proto:
	( cd protos && \
		protoc --go_out=../app/core --go_opt=paths=source_relative \
		--go-grpc_out=../app/core --go-grpc_opt=paths=source_relative \
		./**/*.proto \
	)

gen.web.proto:
	( protoc -I protos/web \
		-I protos/includes/googleapis \
		-I protos/includes/grpc_ecosystem \
		--go_out=./app/web --go_opt=paths=source_relative \
		--go-grpc_out=./app/web --go-grpc_opt=paths=source_relative \
		--openapiv2_out ./openapiv2 --openapiv2_opt logtostderr=true \
		./protos/web/**/*.proto \
	)

build.api.docs:
	( swagger mixin $(find ./openapiv2 -type f -name '*.swagger.json' | tr '\n' ' ') 1>&2 >/dev/null > ./openapiv2/api.swagger.json )

serve.docs:
	docker run -p 8080:8080 \
  	-e SWAGGER_JSON=/openapiv2/api.swagger.json \
  	-v ./openapiv2/:/openapiv2 \
  	swaggerapi/swagger-ui
