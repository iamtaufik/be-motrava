.
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
│   │   ├── google_auth_dto.go
│   │   └── user_dto.go
│   ├── models
│   │   └── user.go
│   ├── port
│   │   └── usecase
│   │       ├── auth_usecase.go
│   │       └── user_usecase.go
│   ├── repository
│   │   └── user_repository.go
│   ├── usecase
│   │   ├── auth_usecase.go
│   │   └── user_usecase.go
│   └── utils
│       └── response
│           └── response.go
├── go.mod
├── go.sum
├── infra
│   ├── database
│   │   └── gorm.go
│   ├── logger
│   │   └── json.go
│   └── repository
│       └── user_repository_gorm.go
├── log.json
└── structure.md

20 directories, 22 files
