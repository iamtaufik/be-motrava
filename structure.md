.
├── Dockerfile
├── Jenkinsfile
├── app
│   ├── app.go
│   ├── handlers
│   │   ├── auth_handler.go
│   │   ├── trip_handler.go
│   │   ├── user_handler.go
│   │   ├── vehicle_handler.go
│   │   └── ws_handler.go
│   ├── middleware
│   │   └── auth.go
│   └── routes
│       └── routes.go
├── cmd
│   └── api
│       └── main.go
├── config
│   └── config.go
├── core
│   ├── dto
│   │   ├── auth_dto.go
│   │   ├── google_auth_dto.go
│   │   ├── trip_dto.go
│   │   ├── user_dto.go
│   │   └── vehicle_dto.go
│   ├── models
│   │   ├── refresh_token.go
│   │   ├── trip.go
│   │   ├── trip_point.go
│   │   ├── user.go
│   │   └── vehicle.go
│   ├── port
│   │   └── usecase
│   │       ├── auth_usecase.go
│   │       ├── trip_usecase.go
│   │       ├── user_usecase.go
│   │       └── vehicle_usecase.go
│   ├── repository
│   │   ├── refresh_token_repository.go
│   │   ├── trip_repository.go
│   │   ├── user_repository.go
│   │   └── vehicle_repository.go
│   ├── usecase
│   │   ├── auth_usecase.go
│   │   ├── trip_usecase.go
│   │   ├── user_usecase.go
│   │   └── vehicle_usecase.go
│   └── utils
│       └── response
│           └── response.go
├── docker-compose.yml
├── go.mod
├── infra
│   ├── database
│   │   ├── gorm.go
│   │   └── redis.go
│   ├── logger
│   │   └── json.go
│   ├── repository
│   │   ├── refresh_token_repository_gorm.go
│   │   ├── trip_repository_gorm.go
│   │   ├── user_repository_gorm.go
│   │   └── vehicle_repository_gorm.go
│   └── ws
│       └── hub.go
├── migrations
│   ├── 001_create_vehicle_table.sql
│   └── 002_create_trip_tables.sql
├── mvp.md
├── structure.md
└── trip-management.postman_collection.json

23 directories, 50 files
