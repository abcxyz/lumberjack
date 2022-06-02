module github.com/abcxyz/lumberjack

go 1.18

require (
	cloud.google.com/go/bigquery v1.32.0
	cloud.google.com/go/compute v1.6.1
	cloud.google.com/go/logging v1.4.2
	github.com/abcxyz/lumberjack/clients/go v0.0.0-00010101000000-000000000000
	github.com/abcxyz/pkg v0.0.0-20220531220657-c0cfdc07493c
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/sethvargo/go-envconfig v0.6.0
	github.com/sethvargo/go-retry v0.2.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/api v0.75.0
	google.golang.org/genproto v0.0.0-20220421151946-72621c1f0bd3
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.0
)

require (
	cloud.google.com/go v0.100.2 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/trace v1.0.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go v1.0.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go/v2 v2.3.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/sdk v1.2.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/net v0.0.0-20220412020605-290c469a71a5 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
	google.golang.org/appengine v1.6.7 // indirect
)

replace github.com/abcxyz/lumberjack/clients/go => ./clients/go
