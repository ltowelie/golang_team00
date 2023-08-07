## Instructions

### ex00

- Install protoc util for your system
https://grpc.io/docs/protoc-installation/

- Install protoc plugins for golang
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

- Generate protobuf and gRPC code, build server and client
```shell
cd ex00
protoc --go_out=./proto --go-grpc_out=./proto ./proto/message.proto
cd ./client
go build
cd ../server
go build
```

- Export envs and run server
```shell
source export
./ex00/server/server
```

- Export envs and run client
```shell
source export
./ex00/client/client
```

### ex01

- Build anomalies detector client
```shell
cd ex01
go build
```

- Generate file with frequencies from ex00 client
```shell
source export
cd ex00/client
go build
./client > ../../ex01/test.data
```

- Small count of anomalies
```shell
cd ex01
< test.data ./anomalyDetection -k 4.6
```

- Bigger count of anomalies
```shell
cd ex01
< test.data ./anomalyDetection -k 4
```

### ex02

- Create db named from export file in PostgreSQL server
```shell
source export
echo $DBNAME 
```

```shell
source export
cd ex02
go build
```

- Generate file with frequencies from ex00 client
```shell
source export
cd ex00/client
go build
./client > ../../ex02/test.data
```

- Run client with anomalies detection and writing its to DB
```shell
source export
cd ex02
< test.data ./anomaliesToDB -k 4.6
```

### ex03

- Create db named from export file in PostgreSQL server
```shell
source export
echo $DBNAME 
```

- Generate protobuf and gRPC code, build server and client
```shell
cd ex03
protoc --go_out=./proto --go-grpc_out=./proto ./proto/message.proto
cd ./client
go build
cd ../server
go build
```

- Export envs and run server
```shell
source export
./ex03/server/server
```

- Export envs and run client
```shell
source export
./ex03/client/client -k 4.6
```