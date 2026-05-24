module github.com/philiaspace/mondaiphi

go 1.22.0

require (
	github.com/lib/pq v1.10.9
	github.com/philiaspace/phi-core v0.0.0
	github.com/philiaspace/phi-exam-domain v0.0.0
	github.com/philiaspace/phi-middleware v0.0.0
	github.com/philiaspace/phi-storage v0.0.0
	github.com/philiaspace/phi-utils v0.0.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/minio-go/v7 v7.0.81 // indirect
	github.com/rs/xid v1.6.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)

replace (
	github.com/philiaspace/phi-core => ../../libs/phi-core
	github.com/philiaspace/phi-exam-domain => ../../libs/phi-exam-domain
	github.com/philiaspace/phi-middleware => ../../libs/phi-middleware
	github.com/philiaspace/phi-storage => ../../libs/phi-storage
	github.com/philiaspace/phi-utils => ../../libs/phi-utils
)
