module github.com/akitasoftware/akita-libs

go 1.18

require (
	github.com/OneOfOne/xxhash v1.2.8
	github.com/akitasoftware/akita-ir v0.0.0-20220630210013-8926783978fe
	github.com/akitasoftware/go-utils v0.0.0-20221207014235-6f4c9079488d
	github.com/akitasoftware/objecthash-proto v0.0.0-20211020004800-9990a7ea5dc0
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.6
	github.com/google/gopacket v1.1.19
	github.com/google/martian/v3 v3.0.1
	github.com/google/uuid v1.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	golang.org/x/exp v0.0.0-20220428152302-39d4317da171
	google.golang.org/protobuf v1.27.1
)

require github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dukex/mixpanel v1.0.1
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/analytics-go/v3 v3.2.1
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace (
	github.com/google/gopacket v1.1.19 => github.com/akitasoftware/gopacket v1.1.18-0.20210730205736-879e93dac35b
	github.com/google/martian/v3 v3.0.1 => github.com/akitasoftware/martian/v3 v3.0.1-0.20210608174341-829c1134e9de
)
