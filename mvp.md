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
| Trip History | ✅ Completed |
| Trip Detail | ✅ Completed |
| Dashboard | ⏳ Todo |
| Vehicle Reminder Service | ✅ Completed |

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

# 6. Trip History ✅

Status: Completed

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

- Pagination (page, limit)
- Search (search by start_address / end_address)
- Filter by Date (date_from, date_to)

## API

- ✅ GET /api/trips

---

# 7. Trip Detail ✅

Status: Completed

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

## API

- ✅ GET /api/trips/:id

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

# 9. Vehicle Reminder Service ✅

Status: Completed

Mengingatkan user ketika kendaraan perlu service berdasarkan jarak tempuh.

## Mekanisme

User membuat service reminder dengan target jarak tempuh (misal 5000 KM).

Akumulasi jarak berasal dari 2 sumber:
- **Trip otomatis**: jarak dari trip yang direkam GPS
- **Input manual**: user bisa input jarak manual (KM) untuk perjalanan yang tidak terekam GPS (misal motor dipinjam orang lain)

Ketika akumulasi sudah mencapai target (≥ 5000 KM), jarak terus terakumulasi (over) hingga user menekan tombol **"Sudah Service"** yang akan me-reset akumulasi ke 0 untuk mulai siklus berikutnya.

## Alur

```
Buat Reminder (target: 5000 KM)
    ↓
Trip GPS (180 KM) + Input Manual (50 KM)
    ↓
Akumulasi = 230 KM
    ↓
... terus bertambah ...
    ↓
Akumulasi ≥ 5000 KM → Progress 100%+
    ↓
User menekan "Sudah Service"
    ↓
Akumulasi di-reset ke 0
    ↓
Siklus baru dimulai
```

## Progress Bar

```
Progress: 230 / 5000 KM (4.6%)
```

Progress bisa > 100% jika user belum menekan "Sudah Service". Contoh: 5200 / 5000 KM (104%).

## Threshold Notifikasi

```
SERVICE_REMINDER_THRESHOLD_PERCENT = 80
```

Ketika progress ≥ 80% dari target, sistem kirim push notification. Hanya terkirim sekali per siklus (dicek via `last_notified_at`).

## Endpoint

### Create Reminder

- ✅ POST /api/vehicles/:id/service-reminders

Body:

```json
{
    "service_name": "Ganti Oli Mesin",
    "interval_km": 5000
}
```

### Get Reminder Progress

- ✅ GET /api/vehicles/:id/service-reminders/:reminderId/progress

Response:

```json
{
    "success": true,
    "message": "service reminder progress retrieved",
    "data": {
        "id": "uuid",
        "service_name": "Ganti Oli Mesin",
        "interval_km": 5000,
        "accumulated_km": 230.5,
        "progress_percent": 4.61,
        "needs_service": false,
        "last_service_at": null,
        "created_at": "2026-07-14T10:00:00Z"
    }
}
```

### List Reminders by Vehicle

- ✅ GET /api/vehicles/:id/service-reminders

### Update Reminder

- ✅ PUT /api/vehicles/:id/service-reminders/:reminderId

### Delete Reminder

- ✅ DELETE /api/vehicles/:id/service-reminders/:reminderId

### Reset After Service

- ✅ POST /api/vehicles/:id/service-reminders/:reminderId/reset

Setelah user melakukan service, reset `accumulated_km` ke 0.

### Add Manual Distance

- ✅ POST /api/vehicles/:id/service-reminders/:reminderId/manual-distance

Untuk input manual jarak perjalanan yang tidak terekam GPS.

Body:

```json
{
    "distance_km": 45.5,
    "note": "Motor dipinjam teman"
}
```

## Sumber Akumulasi Jarak

| Sumber | Cara | Otomatis? |
|--------|------|-----------|
| Trip GPS | Setiap trip selesai, `total_distance` ditambahkan ke akumulasi | ✅ |
| Input Manual | User input via endpoint `manual-distance` | ❌ (manual) |

## Push Notification (FCM)

Menggunakan **Firebase Cloud Messaging (FCM)** — gratis tanpa batas.

### Mekanisme

Dua jalur untuk memastikan notifikasi terkirim:

#### 1. Real-time (event-based) — Primer

Trigger langsung tanpa jeda.

```
Trip selesai (POST /api/trips/:id/end)
    ↓
Input manual distance (POST .../manual-distance)
    ↓
Hitung ulang progress reminder
    ↓
Progress ≥ SERVICE_REMINDER_THRESHOLD_PERCENT (80)?
    ├─ Ya → Cek last_notified_at
    │        ├─ Sudah terkirim di siklus ini? → Skip
    │        └─ Belum → Kirim FCM via Firebase Admin SDK
    └─ Tidak → Skip
```

#### 2. Scheduler (background goroutine) — Fallback

```
go func() {
    ticker := time.NewTicker(30 * time.Minute)
    for range ticker.C {
        Query semua reminder active
            ↓
        Hitung progress (accumulated_km / interval_km * 100)
            ↓
        Progress ≥ SERVICE_REMINDER_THRESHOLD_PERCENT (80)?
            ├─ Ya → Cek last_notified_at
            │        ├─ Sudah terkirim di siklus ini? → Skip
            │        └─ Belum → Kirim FCM
            └─ Tidak → Skip
    }
}()
```

Tabel `M_VEHICLE_SERVICE_REMINDER` punya kolom `notified_at` untuk mencegah duplikasi di kedua jalur.

### Flow

```
Backend (real-time / scheduler)
    ↓
Firebase Admin SDK (HTTP v1 API)
    ↓
Firebase Cloud Messaging (Google Servers)
    ↓
Device (system tray notification)
```

### Payload FCM

```json
{
    "to": "device_fcm_token",
    "notification": {
        "title": "Service Reminder",
        "body": "Ganti Oli Mesin sudah 90% — 4500/5000 KM. Segera service kendaraan Anda."
    },
    "data": {
        "type": "service_reminder",
        "reminder_id": "uuid",
        "vehicle_id": "uuid",
        "service_name": "Ganti Oli Mesin",
        "progress_percent": "90",
        "accumulated_km": "4500",
        "interval_km": "5000"
    }
}
```

### FCM Token

Disimpan di tabel `M_USER_DEVICE` (device_token, platform, created_at) dan dikirim dari Flutter saat login/startup.

---

# Database

## Tables

- M_USER ✅
- M_REFRESH_TOKEN ✅
- M_VEHICLE ✅
- M_TRIP ✅
- M_TRIP_POINT ✅
- M_VEHICLE_SERVICE_REMINDER ✅
- M_MANUAL_DISTANCE_LOG ✅
- M_USER_DEVICE ✅ (FCM token)

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

- ✅ GET /api/trips
- ✅ GET /api/trips/:id
- ✅ POST /api/trips/start
- ✅ POST /api/trips/:id/end
- ✅ WS /api/ws/trip/location?token={jwt}

## Device

- ✅ POST /api/devices/register

## Service Reminder

- ✅ POST /api/vehicles/:vehicleId/service-reminders
- ✅ GET /api/vehicles/:vehicleId/service-reminders
- ✅ GET /api/vehicles/:vehicleId/service-reminders/:reminderId/progress
- ✅ PUT /api/vehicles/:vehicleId/service-reminders/:reminderId
- ✅ DELETE /api/vehicles/:vehicleId/service-reminders/:reminderId
- ✅ POST /api/vehicles/:vehicleId/service-reminders/:reminderId/reset
- ✅ POST /api/vehicles/:vehicleId/service-reminders/:reminderId/manual-distance

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