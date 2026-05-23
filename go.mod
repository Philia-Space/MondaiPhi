module github.com/philiaspace/mondaiphi

go 1.22.0

require (
	github.com/lib/pq v1.10.9
	github.com/philiaspace/phi-core v0.0.0
	github.com/philiaspace/phi-exam-domain v0.0.0
	github.com/philiaspace/phi-middleware v0.0.0
	github.com/philiaspace/phi-utils v0.0.0
)

require github.com/golang-jwt/jwt/v5 v5.3.1 // indirect

replace (
	github.com/philiaspace/phi-core => ../../libs/phi-core
	github.com/philiaspace/phi-exam-domain => ../../libs/phi-exam-domain
	github.com/philiaspace/phi-middleware => ../../libs/phi-middleware
	github.com/philiaspace/phi-storage => ../../libs/phi-storage
	github.com/philiaspace/phi-utils => ../../libs/phi-utils
)
