CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER set_timestamp_users BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE trigger_set_timestamp();

-- Movies: Stores details about the films
CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    duration_minutes INT NOT NULL,
    release_date DATE
);

-- Theaters: The physical cinema buildings
CREATE TABLE theaters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL
);

-- Screens: The individual halls inside a theater
CREATE TABLE screens (
    id SERIAL PRIMARY KEY,
    theater_id INT REFERENCES theaters(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL, -- e.g., "Screen 1", "IMAX"
    total_seats INT NOT NULL
);

-- Seats: The physical seats in a specific screen
CREATE TABLE seats (
    id SERIAL PRIMARY KEY,
    screen_id INT REFERENCES screens(id) ON DELETE CASCADE,
    row_char VARCHAR(2) NOT NULL, -- e.g., 'A', 'B', 'C'
    seat_number INT NOT NULL,     -- e.g., 1, 2, 3
    seat_type VARCHAR(50) DEFAULT 'standard', -- 'standard', 'premium', 'vip'
    UNIQUE(screen_id, row_char, seat_number)
);

-- Shows: A specific movie playing on a specific screen at a specific time
CREATE TABLE shows (
    id SERIAL PRIMARY KEY,
    movie_id INT REFERENCES movies(id) ON DELETE CASCADE,
    screen_id INT REFERENCES screens(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    base_price DECIMAL(10, 2) NOT NULL
);

-- Bookings: The main reservation record for a user
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    show_id INT REFERENCES shows(id) ON DELETE CASCADE,
    total_price DECIMAL(10, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'confirmed', 'cancelled'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Show Seats (The crucial table for concurrency!): 
-- Tracks the status of every single seat for a specific show.
CREATE TABLE show_seats (
    id SERIAL PRIMARY KEY,
    show_id INT REFERENCES shows(id) ON DELETE CASCADE,
    seat_id INT REFERENCES seats(id) ON DELETE CASCADE,
    booking_id INT REFERENCES bookings(id) ON DELETE SET NULL,
    status VARCHAR(50) DEFAULT 'available', -- 'available', 'locked', 'booked'
    locked_until TIMESTAMP, -- Used for our 5-minute checkout timer!
    UNIQUE(show_id, seat_id)
);