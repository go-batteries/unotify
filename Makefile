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
		--openapiv2_out ./openapiv2 --openapiv2_opt logtostderr=true \
		./protos/web/**/*.proto \
	)

go.test:
	go test -timeout=20s ./...

serve.docs: gen.web.proto
	( bash ./build.docs.sh )

run.server:
	go run -race cmd/server/main.go
