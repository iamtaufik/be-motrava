CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE "M_VEHICLE" (
    "id"          UUID         NOT NULL DEFAULT uuid_generate_v4(),
    "user_id"     UUID         NOT NULL,
    "vehicle_name" VARCHAR(255) NOT NULL,
    "plate_number" VARCHAR(50)  NOT NULL,
    "brand"       VARCHAR(100) NOT NULL,
    "model"       VARCHAR(100) NOT NULL,
    "vehicle_type" VARCHAR(20)  NOT NULL DEFAULT 'CAR',
    "color"       VARCHAR(50)  NOT NULL,
    "year"        INT          NULL,
    "photo"       TEXT         NULL,
    "is_default"  BOOLEAN      NOT NULL DEFAULT false,
    "created_at"  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "updated_at"  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    "deleted_at"  TIMESTAMPTZ  NULL,

    CONSTRAINT "M_VEHICLE_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "fk_M_VEHICLE_user" FOREIGN KEY ("user_id") REFERENCES "M_USER"("id") ON DELETE CASCADE
);

CREATE INDEX "idx_M_VEHICLE_user_id" ON "M_VEHICLE"("user_id");
CREATE INDEX "idx_M_VEHICLE_deleted_at" ON "M_VEHICLE"("deleted_at");
