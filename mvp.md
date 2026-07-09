# 🚗 Motrava MVP Roadmap

> MVP Version: v1.0
>
> Backend : Golang Fiber
> Database : PostgreSQL + PostGIS
> Cache : Redis
> Mobile : Flutter
> Realtime : WebSocket

---

# Progress

| Feature | Status |
|----------|--------|
| Authentication | ✅ Completed |
| Vehicle Management | ✅ Completed |
| Trip Tracking | ✅ Completed |
| Live Tracking | ✅ Completed |
| Trip Summary | ✅ Completed |
| Trip History | ⏳ Todo |
| Trip Detail | ⏳ Todo |
| Dashboard | ⏳ Todo |

---

# 1. Authentication ✅

Status: Completed

## Features

- Register with Email
- Login with Email
- Login with Google
- Refresh Token
- Logout
- User Profile
- JWT Authentication

---

# 2. Vehicle Management ✅

Status: Completed

## Description

Satu user dapat memiliki banyak kendaraan.

## Features

- Create Vehicle
- Update Vehicle
- Delete Vehicle
- Get Vehicle List
- Set Default Vehicle

## Vehicle Information

- Vehicle Name
- Plate Number
- Brand
- Model
- Vehicle Type
- Color
- Year (Optional)
- Photo (Optional)

Example

```
Toyota Avanza
B 1234 ABC

Honda Beat
AG 4321 YY
```

---

# 3. Start Trip ✅

Status: Completed

## Description

User memilih kendaraan kemudian menekan tombol Start.

Backend membuat Trip baru dan aplikasi mulai mengirim lokasi.

## Flow

```
Select Vehicle

↓

Start Trip

↓

Backend Create Trip

↓

GPS Tracking Started
```

## GPS Payload

```json
{
    "latitude": -7.8123,
    "longitude": 110.3642,
    "speed": 43.5,
    "heading": 270,
    "accuracy": 5,
    "altitude": 120,
    "battery": 87,
    "timestamp": "2026-07-09T10:00:00Z"
}
```

Tracking Interval

- Every 1 second

or

- Every 5 meter movement

---

# 4. Live Tracking ✅

Status: Completed

## Description

Selama perjalanan aplikasi terus mengirim lokasi.

Backend

- Save GPS Point
- Broadcast via WebSocket
- Update Current Position

Flow

```
Flutter

↓

WebSocket

↓

Fiber

↓

Redis

↓

Dashboard
```

---

# 5. End Trip ✅

Status: Completed

## Description

Ketika user menekan Stop.

Backend menghitung statistik perjalanan.

## Statistics

- Total Distance
- Duration
- Moving Time
- Idle Time
- Average Speed
- Maximum Speed
- Start Location
- End Location

Kemudian status Trip menjadi Completed.

---

# 6. Trip History

Status: Todo

Menampilkan daftar perjalanan.

Example

```
Home → Office

Distance
18.3 km

Duration
32 Minutes

Average Speed
45 km/h

Date
09 July 2026
```

Features

- Pagination
- Search
- Filter by Date

---

# 7. Trip Detail

Status: Todo

Menampilkan detail perjalanan.

## Information

- Route Polyline
- Start Marker
- Finish Marker
- Distance
- Duration
- Average Speed
- Maximum Speed
- Moving Time
- Idle Time

Future

- Elevation Chart
- Speed Chart

---

# 8. Dashboard

Status: Todo

Menampilkan ringkasan aktivitas user.

## Statistics

- Total Trips
- Total Distance
- Total Driving Time
- Average Speed
- Maximum Speed
- This Week Distance
- This Month Distance

---

# Database

## Tables

- M_USER ✅
- M_REFRESH_TOKEN ✅
- M_VEHICLE ✅
- M_TRIP ✅
- M_TRIP_POINT ✅

---

# API

## Authentication

- ✅ POST /api/auth/register
- ✅ POST /api/auth/login
- ✅ POST /api/auth/refresh
- ✅ GET /api/auth/me
- ✅ GET /api/auth/google/login
- ✅ GET /api/auth/google/callback
- ✅ POST /api/auth/google/mobile

## Vehicle

- ✅ POST /api/vehicles
- ✅ GET /api/vehicles
- ✅ GET /api/vehicles/:id
- ✅ PUT /api/vehicles/:id
- ✅ DELETE /api/vehicles/:id
- ✅ PUT /api/vehicles/:id/default

## Trip

- ✅ POST /api/trips/start
- ✅ POST /api/trips/:id/end
- ✅ WS /api/ws/trip/location?token={jwt}

---

# Future MVP (v2)

- Geofence
- Driving Score
- Overspeed Detection
- Push Notification
- Share Trip
- Community
- Leaderboard
- Fuel Estimation
- OBD-II Integration
- Fleet Management
- Route Replay
- Trip Export (GPX/CSV)

---

# Tech Stack

Backend

- Golang Fiber
- GORM
- PostgreSQL
- PostGIS
- Redis
- WebSocket

Frontend

- Flutter

Maps

- OpenStreetMap
- Flutter Map

Routing

- OSRM

Reverse Geocoding

- Nominatim

Storage

- MinIO