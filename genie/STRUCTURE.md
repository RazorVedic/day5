Directory Structure for UPI Switch in Golang

Instructions:
1. STRICTLY FOLLOW THIS DIRECTORY STRUCTURE: Do not deviate from the provided structure.
2. DO NOT ADD '####' PREFIX TO GENERATED FILE PATHS: Ensure generated file paths do not have the '####' prefix.

Root directory path: github.com/razorpay/upi-switch/

Please strictly follow the below directory structure in golang

```
upi-switch/
 ├── cmd/
 │   └── <aggregate-name>/
 │       ├── config/
 │       │   └── config.go
 │       └── main.go
 │   └── <aggregate-name>-migration/
 │       └── main.go
 ├── config/
 │   └── <aggregate-name>/
 │       ├── default.go
 ├── internal/
 │   ├── <aggregate-name>/
 │   │   ├── migrations/
 │   │   │   ├── 20240508180733_create_<model1>.go
 │   │   │   ├── 20240508180733_create_<model2>.go
 │   │   ├── service/
 │   │   │   ├── core/
 │   │   │   │   ├── model/
 │   │   │   │   │   └── model.go
 │   │   │   │   ├── repo/
 │   │   │   │   │   └── repo.go
 │   │   │   │   └── core.go
 │   │   │   ├── service.go
 │   │   │   ├── option.go
 │   │   ├── validator/
 │   │   │   ├── validator.go
 │   │   │   ├── validator_test.go
 │   │   ├── server.go
 │   │   ├── server_test.go
 │   ├── topics/
 │   │   ├── <aggregate-name>.go
 ├── pkg/
 │   ├── pubsub/
 │   │   ├── pub/
 │   │   │   ├── pub.go
 │   │   ├── sub/
 │   │   │   ├── sub.go
 │   ├── storage/
 │   │   ├── storage.go
 │   │   ├── sql/
 │   │   │   ├── gorm.go
 │   ├── observability/
 │   │   ├── observability.go 
 │   ├── server/
 │   │   ├── interceptor/
 │   │   │   ├── interceptors.go
 │   │   ├── server.go
 │   ├── configloader/
 │   │   ├── configloader.go 
 ├── proto/
 │   ├── switch/v1/
 │   │   ├── <aggregate-name>/
 │   │   │   ├── <aggregate-name>.proto
 │   │   │   ├── event/
 │   │   │   │   ├── <aggregate-name>.proto
 │   │   ├── event.proto
 ├── rpc/
```

Details:
- <aggregate-name>: Replace with the specific aggregate name.
- <model1> and <model2>: Replace with specific model names.
- Each directory and file must be created as specified.

Notes:
- Pay special attention to the hierarchical structure.
- Ensure correct file naming conventions.
- Maintain consistency across all generated paths