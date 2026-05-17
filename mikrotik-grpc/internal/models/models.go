package models

import (
	"time"
)

// Package represents a service plan in the database.
type Package struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	RateLimit string    `db:"rate_limit" json:"rate_limit"`
	Price     float64   `db:"price" json:"price"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// User represents a subscriber in the database.
type User struct {
	ID         string    `db:"id" json:"id"`
	Username   string    `db:"username" json:"username"`
	Password   string    `db:"password" json:"-"` // Omit password from JSON responses
	PackageID  string    `db:"package_id" json:"package_id"`
	Status     string    `db:"status" json:"status"`
	ExpiryDate time.Time `db:"expiry_date" json:"expiry_date"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`

	// This field is joined from the packages table for RADIUS responses
	PackageRateLimit string `db:"rate_limit" json:"package_rate_limit,omitempty"`
}

// Session represents an accounting session in the database.
type Session struct {
	ID               int64     `db:"id" json:"id"`
	SessionID        string    `db:"session_id" json:"session_id"`
	Username         string    `db:"username" json:"username"`
	NASIPAddress     string    `db:"nas_ip_address" json:"nas_ip_address"`
	CallingStationID string    `db:"calling_station_id" json:"calling_station_id"`
	SessionStartTime time.Time `db:"session_start_time" json:"session_start_time"`
	SessionStopTime  time.Time `db:"session_stop_time" json:"session_stop_time,omitempty"`
	SessionTotalTime int       `db:"session_total_time" json:"session_total_time"`
	InputOctets      int64     `db:"input_octets" json:"input_octets"`
	OutputOctets     int64     `db:"output_octets" json:"output_octets"`
	TerminateCause   string    `db:"terminate_cause" json:"terminate_cause,omitempty"`
}