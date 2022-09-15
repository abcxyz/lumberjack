module github.com/abcxyz/lumberjack

go 1.19

require (
	cloud.google.com/go/bigquery v1.32.0
	cloud.google.com/go/compute v1.6.1
	cloud.google.com/go/logging v1.4.2
	github.com/abcxyz/jvs v0.0.2-0.20220915004312-db26d444b9ae
	github.com/abcxyz/pkg v0.0.0-20220719233420-62c7b76c10e9
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/sethvargo/go-envconfig v0.8.0
	github.com/sethvargo/go-retry v0.2.3
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0
	golang.org/x/oauth2 v0.0.0-20220524215830-622c5d57e401
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	google.golang.org/api v0.82.0
	google.golang.org/genproto v0.0.0-20220602131408-e326c6e8e9c8
	google.golang.org/grpc v1.47.0
	google.golang.org/protobuf v1.28.0
)

require (
	cloud.google.com/go v0.102.0 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/kms v1.4.0 // indirect
	cloud.google.com/go/trace v1.0.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go v1.0.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.0.0 // indirect
	github.com/abcxyz/jvs/client-lib/go v0.0.0-20220915005309-051e121fe9e7 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/goccy/go-json v0.9.7 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.2 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.0.3 // indirect
	github.com/lestrrat-go/option v1.0.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.opentelemetry.io/otel v1.7.0 // indirect
	go.opentelemetry.io/otel/sdk v1.2.0 // indirect
	go.opentelemetry.io/otel/trace v1.7.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220427172511-eb4f295cb31f // indirect
	golang.org/x/net v0.0.0-20220531201128-c960675eff93 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df // indirect
	google.golang.org/appengine v1.6.7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/abcxyz/lumberjack/clients/go => ./clients/go
