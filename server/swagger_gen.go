package server

// IMPORTANT! MUST USE go_swagger_version=0.27.0 to generate.
// reference website: https://github.com/go-swagger/go-swagger/releases/tag/v0.27.0
//go:generate go run ./delgensource/main.go
//go:generate swagger generate model --target . --spec ../assets/mesage/openapi.yaml --model-package=models
//go:generate swagger generate server --target swagger --name mui --spec ../swagger.yml --principal models.Principal
//go:generate swagger generate model --target . --spec ../assets/mes/services.swagger.json --model-package=mes
