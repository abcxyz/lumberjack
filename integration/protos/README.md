To generate Go code:

```
protoc -I. --go_out=. --go-grpc_out=. \
  --go_opt=module=github.com/abcxyz/lumberjack \
  --go-grpc_opt=module=github.com/abcxyz/lumberjack \
  integration/protos/talker.proto
```