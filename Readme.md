# k8s_monitoring_golang

## Real time monitoring of Kubernetes components using [Golang]

```go
file structure for this project : 

├── Readme.md
├── go.mod
├── go.sum
├── importConfig.go
├── k8sApi.go
├── main.go
└── web-socket.go

run the server using the below command
go run *.go

(/importConfig) endpoint is exposed which requires the yaml file for connecting to kubernestes cluster

once the connection is established every 10sec the following metrics will be provided:
1. list of pods and their respective status
2. cpu and memory usage of the containers in the pods
3. restart counts of the pods

```