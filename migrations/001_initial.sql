CREATE TABLE IF NOT EXISTS roads (id TEXT PRIMARY KEY, name TEXT NOT NULL, code TEXT NOT NULL UNIQUE, level TEXT NOT NULL, speed_limit INTEGER NOT NULL, lanes INTEGER NOT NULL, length_meters NUMERIC NOT NULL, status TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS intersections (id TEXT PRIMARY KEY, name TEXT NOT NULL, code TEXT NOT NULL UNIQUE, latitude NUMERIC NOT NULL, longitude NUMERIC NOT NULL, level TEXT NOT NULL, status TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS timing_plans (id TEXT PRIMARY KEY, intersection_id TEXT NOT NULL, name TEXT NOT NULL, cycle INTEGER NOT NULL, offset_seconds INTEGER NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL);
CREATE TABLE IF NOT EXISTS traffic_readings (id TEXT PRIMARY KEY, intersection_id TEXT NOT NULL, sensor_id TEXT NOT NULL, lane TEXT NOT NULL, volume INTEGER NOT NULL, speed NUMERIC NOT NULL, queue NUMERIC NOT NULL, occupancy NUMERIC NOT NULL, recorded_at TIMESTAMPTZ NOT NULL);
CREATE TABLE IF NOT EXISTS traffic_events (id TEXT PRIMARY KEY, type TEXT NOT NULL, level TEXT NOT NULL, title TEXT NOT NULL, status TEXT NOT NULL, road_id TEXT, intersection_id TEXT, started_at TIMESTAMPTZ NOT NULL, description TEXT NOT NULL, source TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_traffic_readings_intersection_time ON traffic_readings(intersection_id, recorded_at);
CREATE INDEX IF NOT EXISTS idx_events_status ON traffic_events(status);
