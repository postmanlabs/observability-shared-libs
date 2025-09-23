module github.com/akitasoftware/akita-libs

go 1.18

require (
	github.com/OneOfOne/xxhash v1.2.8
	github.com/akitasoftware/akita-ir v0.0.0-20241213050034-057d7b6097e8
	github.com/akitasoftware/go-utils v0.0.0-20221207014235-6f4c9079488d
	github.com/akitasoftware/objecthash-proto v0.0.0-20211020004800-9990a7ea5dc0
	github.com/amplitude/analytics-go v1.0.1
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.8
	github.com/google/gopacket v1.1.19
	github.com/google/uuid v1.3.0
	github.com/iancoleman/strcase v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/segmentio/analytics-go/v3 v3.3.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/exp v0.0.0-20240416160154-fe59bbe5cc7f
	google.golang.org/protobuf v1.27.1
)

require (
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/backo-go v1.0.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/google/gopacket v1.1.19 => github.com/akitasoftware/gopacket v1.1.18-0.20240820200020-7289ae956f70
