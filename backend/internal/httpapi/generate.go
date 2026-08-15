package httpapi

// A geração do servidor Go a partir da spec. `task generate` chama
// `go generate ./...`, então este é o único lugar que define o comando.
//
//go:generate go tool oapi-codegen -config ../../oapi-codegen.yaml -o gen/api.gen.go ../../../api/openapi.yaml
