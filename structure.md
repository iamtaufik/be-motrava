.
├── Dockerfile
├── Jenkinsfile
├── app
│   ├── app.go
│   ├── handlers
│   │   ├── auth_handler.go
│   │   └── user_handler.go
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
│   │   └── user_dto.go
│   ├── models
│   │   ├── refresh_token.go
│   │   └── user.go
│   ├── port
│   │   └── usecase
│   │       ├── auth_usecase.go
│   │       └── user_usecase.go
│   ├── repository
│   │   ├── refresh_token_repository.go
│   │   └── user_repository.go
│   ├── usecase
│   │   ├── auth_usecase.go
│   │   └── user_usecase.go
│   └── utils
│       └── response
│           └── response.go
├── docker-compose.yml
├── go.mod
├── go.sum
├── infra
│   ├── database
│   │   └── gorm.go
│   ├── logger
│   │   └── json.go
│   └── repository
│       ├── refresh_token_repository_gorm.go
│       └── user_repository_gorm.go
├── log.json
├── structure.md
└── tmp
    ├── build-errors.log
    └── main

21 directories, 31 files
