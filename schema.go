package main

// Booking represents a single booking entity
type Booking struct {
	ID       string `json:"id"`
	Date     string `json:"date"`
	User     string `json:"user"`
	UserName string `json:"user_name,omitempty"`
	Area     string `json:"area"`
	AreaData Area   `json:"area_data"`
	AreaRef  string `json:"area_ref,omitempty"`
}

// AddBookingRequest represents a request object for creating a new booking at one or more dates
type AddBookingRequest struct {
	Area  string   `json:"area"`
	Dates []string `json:"dates"`
	Start string   `json:"start"`
	End   string   `json:"end"`
}

// Bookings represents a list of Booking items
type Bookings struct {
	Bookings []Booking `json:"bookings"`
}

// Area represents a single area entity
type Area struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	Capacity uint16 `json:"capacity"`
	Usage    uint16 `json:"usage"`
	Location string `json:"location"`
	Type     string `json:"type"`
}

// Areas represents a list of Area items
type Areas struct {
	Areas []Area `json:"areas"`
}

// Forecast
type Forecast struct {
	CreatedAt string         `json:"created_at"`
	Bookings  []ForecastItem `json:"bookings"`
	Area      Area           `json:"area"`
}

// ForecastItem
type ForecastItem struct {
	Date           string `json:"date"`
	BookedSeats    uint16 `json:"booked_seats"`
	BookedByMyself bool   `json:"booked_by_myself"`
}

// ErrorResponse
type ErrorResponse struct {
	Code   int      `json:"code"`
	Errors []string `json:"errors"`
}

// SuccessResponse
type SuccessResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// BacktracingResponse
type BacktracingResponse struct {
	For       string            `json:"for"`
	CreatedAt string            `json:"created_at"`
	NotBefore string            `json:"not_before"`
	NotAfter  string            `json:"not_after"`
	Data      []BacktracingItem `json:"data"`
}

// BacktracingItems
type BacktracingItem struct {
	Date     string `json:"date"`
	Email    string `json:"email"`
	AreaName string `json:"area_name"`
}
