package main
//go:generate go tool oapi-codegen -config ../api/http_config.yaml ../api/openapi.yaml
//go:generate go tool oapi-codegen -config ../api/gin_config.yaml ../api/openapi.yaml
//go:generate go tool oapi-codegen -config ../api/schemas_config.yaml ../api/schemas.yaml