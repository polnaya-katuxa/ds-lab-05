package cars_service

//go:generate oapi-codegen --config=config.yaml ../../../../../../cars-service/api/openapi/cars-service.yaml

//go:generate mockery --all --with-expecter --exported --output mocks/
