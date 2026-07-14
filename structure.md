.
├── Dockerfile
├── Jenkinsfile
├── app
│   ├── app.go
│   ├── handlers
│   │   ├── auth_handler.go
│   │   ├── device_handler.go
│   │   ├── service_reminder_handler.go
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
│   │   ├── device_dto.go
│   │   ├── google_auth_dto.go
│   │   ├── service_reminder_dto.go
│   │   ├── trip_dto.go
│   │   ├── user_dto.go
│   │   └── vehicle_dto.go
│   ├── models
│   │   ├── manual_distance_log.go
│   │   ├── refresh_token.go
│   │   ├── service_reminder.go
│   │   ├── trip.go
│   │   ├── trip_point.go
│   │   ├── user.go
│   │   ├── user_device.go
│   │   └── vehicle.go
│   ├── port
│   │   ├── fcm
│   │   │   └── fcm.go
│   │   └── usecase
│   │       ├── auth_usecase.go
│   │       ├── device_usecase.go
│   │       ├── service_reminder_usecase.go
│   │       ├── trip_usecase.go
│   │       ├── user_usecase.go
│   │       └── vehicle_usecase.go
│   ├── repository
│   │   ├── refresh_token_repository.go
│   │   ├── service_reminder_repository.go
│   │   ├── trip_repository.go
│   │   ├── user_device_repository.go
│   │   ├── user_repository.go
│   │   └── vehicle_repository.go
│   ├── usecase
│   │   ├── auth_usecase.go
│   │   ├── device_usecase.go
│   │   ├── reminder_notifier.go
│   │   ├── service_reminder_usecase.go
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
│   ├── fcm
│   │   ├── client.go
│   │   └── noop.go
│   ├── logger
│   │   └── json.go
│   ├── repository
│   │   ├── refresh_token_repository_gorm.go
│   │   ├── service_reminder_repository_gorm.go
│   │   ├── trip_repository_gorm.go
│   │   ├── user_device_repository_gorm.go
│   │   ├── user_repository_gorm.go
│   │   └── vehicle_repository_gorm.go
│   └── ws
│       └── hub.go
├── migrations
│   ├── 001_create_vehicle_table.sql
│   ├── 002_create_trip_tables.sql
│   └── 003_create_service_reminder_tables.sql
├── mvp.md
├── structure.md
└── trip-management.postman_collection.json

23 directories, 50 files
