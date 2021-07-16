protoc -I protos/ protos/file_streamer.proto --go_out=protos/file_streamer --go_opt=paths=source_relative \
  --go-grpc_out=protos/file_streamer --go-grpc_opt=paths=source_relative
