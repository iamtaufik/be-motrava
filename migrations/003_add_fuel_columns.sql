-- Add fuel efficiency column to vehicle table
ALTER TABLE "M_VEHICLE"
    ADD COLUMN "fuel_efficiency_km_per_liter" DOUBLE PRECISION NULL;

-- Add fuel consumed column to trip table
ALTER TABLE "M_TRIP"
    ADD COLUMN "fuel_consumed" DOUBLE PRECISION NULL;
