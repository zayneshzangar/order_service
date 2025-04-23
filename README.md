go get github.com/rs/cors
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/handlers



// TEST
go install go.uber.org/mock/mockgen@latest

mockgen -source=internal/repository/order_repository.go -destination=internal/repository/mocks/mock_order_repository.go -package=mocks
mockgen -source=internal/delivery/grpcclient/interfaces.go -destination=internal/delivery/grpcclient/mocks/mock_service_client.go -package=mocks
mockgen -source=internal/service/service.go -destination=internal/service/mocks/mock_order_service.go -package=mocks

go get go.uber.org/mock/gomock
go mod tidy

go clean -testcache
go test -v ./internal/service
