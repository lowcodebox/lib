# LogBox Client

## Описание
Предназначен работы с сервисом logbox

## Генерация

- переходим в директорию pkg/model
- выполните команду (для SDK Go)

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./model.proto`
