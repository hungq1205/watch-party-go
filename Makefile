gen:
	@protoc \
		--proto_path=protobuf "protobuf/users.proto3" \
		--go_out=services/common/genproto/users --go_opt=paths=source_relative \
		--go-grpc_out=services/common/genproto/users
		--go-grpc_out=paths=source_relative

		