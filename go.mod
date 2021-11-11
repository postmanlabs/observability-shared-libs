module github.com/akitasoftware/akita-libs

go 1.15

require (
	github.com/OneOfOne/xxhash v1.2.8
	github.com/akitasoftware/akita-ir v0.0.0-20211111012430-2a7dcb20a144
	github.com/akitasoftware/objecthash-proto v0.0.0-20211020004800-9990a7ea5dc0
	github.com/andreyvit/diff v0.0.0-20170406064948-c7f18ee00883
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.5.0
	github.com/google/go-cmp v0.5.6
	github.com/google/gopacket v1.1.19
	github.com/google/martian/v3 v3.0.1
	github.com/google/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/stretchr/testify v1.7.0
	google.golang.org/protobuf v1.27.1
)

replace (
	github.com/google/gopacket v1.1.19 => github.com/akitasoftware/gopacket v1.1.18-0.20210730205736-879e93dac35b
	github.com/google/martian/v3 v3.0.1 => github.com/akitasoftware/martian/v3 v3.0.1-0.20210608174341-829c1134e9de
)
