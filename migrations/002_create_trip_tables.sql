CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- M_TRIP: menyimpan data perjalanan
CREATE TABLE "M_TRIP" (
    "id"              UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "user_id"         UUID         NOT NULL,
    "vehicle_id"      UUID         NOT NULL,
    "start_time"      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "end_time"        TIMESTAMPTZ  NULL,
    "start_latitude"  DOUBLE PRECISION NOT NULL DEFAULT 0,
    "start_longitude" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "end_latitude"    DOUBLE PRECISION NULL,
    "end_longitude"   DOUBLE PRECISION NULL,
    "start_address"   TEXT         NULL,
    "end_address"     TEXT         NULL,
    "total_distance"  DOUBLE PRECISION NOT NULL DEFAULT 0,
    "duration"        INTEGER      NOT NULL DEFAULT 0,
    "moving_time"     INTEGER      NOT NULL DEFAULT 0,
    "idle_time"       INTEGER      NOT NULL DEFAULT 0,
    "average_speed"   DOUBLE PRECISION NOT NULL DEFAULT 0,
    "maximum_speed"   DOUBLE PRECISION NOT NULL DEFAULT 0,
    "status"          VARCHAR(20)  NOT NULL DEFAULT 'ONGOING',
    "created_at"      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "updated_at"      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "deleted_at"      TIMESTAMPTZ  NULL,

    CONSTRAINT "M_TRIP_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_TRIP_user" FOREIGN KEY ("user_id")
        REFERENCES "M_USER"("id") ON DELETE CASCADE,
    CONSTRAINT "fk_M_TRIP_vehicle" FOREIGN KEY ("vehicle_id")
        REFERENCES "M_VEHICLE"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_TRIP_user_id" ON "M_TRIP"("user_id");
CREATE INDEX "idx_M_TRIP_status" ON "M_TRIP"("status");
CREATE INDEX "idx_M_TRIP_deleted_at" ON "M_TRIP"("deleted_at");

-- M_TRIP_POINT: menyimpan titik GPS selama perjalanan
CREATE TABLE "M_TRIP_POINT" (
    "id"          UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "trip_id"     UUID         NOT NULL,
    "latitude"    DOUBLE PRECISION NOT NULL,
    "longitude"   DOUBLE PRECISION NOT NULL,
    "speed"       DOUBLE PRECISION NOT NULL DEFAULT 0,
    "heading"     DOUBLE PRECISION NOT NULL DEFAULT 0,
    "accuracy"    DOUBLE PRECISION NOT NULL DEFAULT 0,
    "altitude"    DOUBLE PRECISION NOT NULL DEFAULT 0,
    "battery"     INTEGER      NOT NULL DEFAULT 0,
    "recorded_at" TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT "M_TRIP_POINT_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_TRIP_POINT_trip" FOREIGN KEY ("trip_id")
        REFERENCES "M_TRIP"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_TRIP_POINT_trip_id" ON "M_TRIP_POINT"("trip_id");
CREATE INDEX "idx_M_TRIP_POINT_recorded_at" ON "M_TRIP_POINT"("recorded_at");
