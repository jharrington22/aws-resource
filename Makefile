# Ensure go modules are enabled:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0

# Unset GOFLAG for CI and ensure we've got nothing accidently set
unexport GOFLAGS

.PHONY: aws-resources 
aws-resources:
	go build -o aws-resource ./cmd/aws-resource/main.go

.PHONY: test
test:
	go test ./...

.PHONY: install
install:
	go install ./cmd/aws-resource/main.go

.PHONY: fmt
fmt:
	gofmt -s -l -w cmd pkg

.PHONY: lint
lint:
	golangci-lint run --timeout 5m0s

.PHONY: clean
clean:
	rm -rf \
		aws-resource \
		*-darwin-amd64 \
		*-linux-amd64 \
		*-windows-amd64 \
		*.sha256 \
		$(NULL)

.PHONY: generate

mocks:
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/iamapi.go github.com/aws/aws-sdk-go/service/iam/iamiface IAMAPI
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/organaztionsapi.go github.com/aws/aws-sdk-go/service/organizations/organizationsiface OrganizationsAPI
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/stsapi.go github.com/aws/aws-sdk-go/service/sts/stsiface STSAPI
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/cloudformationapi.go github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface CloudFormationAPI
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/ec2api.go github.com/aws/aws-sdk-go/service/ec2/ec2iface EC2API
	mockgen --build_flags=--mod=mod -package mocks -destination=pkg/aws/mocks/servicequotasapi.go github.com/aws/aws-sdk-go/service/servicequotas/servicequotasiface ServiceQuotasAPI
	mockgen --build_flags=--mod=mod -package mocks -destination=cmd/create/idp/mocks/identityprovider.go -source=cmd/create/idp/cmd.go IdentityProvider
