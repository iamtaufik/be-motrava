.
├── Dockerfile
├── Jenkinsfile
├── app
│   ├── app.go
│   ├── handlers
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   └── vehicle_handler.go
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
│   │   ├── user_dto.go
│   │   └── vehicle_dto.go
│   ├── models
│   │   ├── refresh_token.go
│   │   ├── user.go
│   │   └── vehicle.go
│   ├── port
│   │   └── usecase
│   │       ├── auth_usecase.go
│   │       ├── user_usecase.go
│   │       └── vehicle_usecase.go
│   ├── repository
│   │   ├── refresh_token_repository.go
│   │   ├── user_repository.go
│   │   └── vehicle_repository.go
│   ├── usecase
│   │   ├── auth_usecase.go
│   │   ├── user_usecase.go
│   │   └── vehicle_usecase.go
│   └── utils
│       └── response
│           └── response.go
├── docker-compose.yml
├── go.mod
├── infra
│   ├── database
│   │   └── gorm.go
│   ├── logger
│   │   └── json.go
│   └── repository
│       ├── refresh_token_repository_gorm.go
│       ├── user_repository_gorm.go
│       └── vehicle_repository_gorm.go
├── mvp.md
└── structure.md

21 directories, 36 files
