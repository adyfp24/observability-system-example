package dto

type RegisterRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,min=9,max=15"`
	Password    string `json:"password" validate:"required,min=6"`
	Role        string `json:"role" validate:"required,oneof=customer driver admin"`
	VehicleType string `json:"vehicle_type" validate:"required_if=Role driver"` // car_premium, car_regular, bike
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required"`
	Password    string `json:"password" validate:"required"`
}

type LoginResponse struct {
	UserID uint64 `json:"user_id"`
	Token  string `json:"token"`
	Role   string `json:"role"`
}

type UserProfile struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
}
