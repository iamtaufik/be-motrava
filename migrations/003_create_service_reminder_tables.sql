CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- M_VEHICLE_SERVICE_REMINDER: menyimpan pengingat service kendaraan
CREATE TABLE "M_VEHICLE_SERVICE_REMINDER" (
    "id"               UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "vehicle_id"       UUID         NOT NULL,
    "user_id"          UUID         NOT NULL,
    "service_name"     VARCHAR(255) NOT NULL,
    "interval_km"      DOUBLE PRECISION NOT NULL,
    "accumulated_km"   DOUBLE PRECISION NOT NULL DEFAULT 0,
    "notified_at"      TIMESTAMPTZ  NULL,
    "last_service_at"  TIMESTAMPTZ  NULL,
    "is_active"        BOOLEAN      NOT NULL DEFAULT TRUE,
    "created_at"       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "updated_at"       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "deleted_at"       TIMESTAMPTZ  NULL,

    CONSTRAINT "M_VEHICLE_SERVICE_REMINDER_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_VEHICLE_SERVICE_REMINDER_vehicle" FOREIGN KEY ("vehicle_id")
        REFERENCES "M_VEHICLE"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_M_VEHICLE_SERVICE_REMINDER_user" FOREIGN KEY ("user_id")
        REFERENCES "M_USER"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_VEHICLE_SERVICE_REMINDER_vehicle_id" ON "M_VEHICLE_SERVICE_REMINDER"("vehicle_id");
CREATE INDEX "idx_M_VEHICLE_SERVICE_REMINDER_user_id" ON "M_VEHICLE_SERVICE_REMINDER"("user_id");
CREATE INDEX "idx_M_VEHICLE_SERVICE_REMINDER_is_active" ON "M_VEHICLE_SERVICE_REMINDER"("is_active");
CREATE INDEX "idx_M_VEHICLE_SERVICE_REMINDER_deleted_at" ON "M_VEHICLE_SERVICE_REMINDER"("deleted_at");

-- M_MANUAL_DISTANCE_LOG: menyimpan riwayat input jarak manual
CREATE TABLE "M_MANUAL_DISTANCE_LOG" (
    "id"          UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "reminder_id" UUID         NOT NULL,
    "distance_km" DOUBLE PRECISION NOT NULL,
    "note"        TEXT         NULL,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT "M_MANUAL_DISTANCE_LOG_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_MANUAL_DISTANCE_LOG_reminder" FOREIGN KEY ("reminder_id")
        REFERENCES "M_VEHICLE_SERVICE_REMINDER"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_MANUAL_DISTANCE_LOG_reminder_id" ON "M_MANUAL_DISTANCE_LOG"("reminder_id");

-- M_USER_DEVICE: menyimpan FCM token perangkat user
CREATE TABLE "M_USER_DEVICE" (
    "id"           UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "user_id"      UUID         NOT NULL,
    "device_token" TEXT         NOT NULL,
    "platform"     VARCHAR(20)  NOT NULL DEFAULT 'android',
    "created_at"   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "updated_at"   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT "M_USER_DEVICE_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_USER_DEVICE_user" FOREIGN KEY ("user_id")
        REFERENCES "M_USER"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_USER_DEVICE_user_id" ON "M_USER_DEVICE"("user_id");
CREATE UNIQUE INDEX "idx_M_USER_DEVICE_token" ON "M_USER_DEVICE"("device_token");
