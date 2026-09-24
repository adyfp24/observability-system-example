package dto

type TopUpRequest struct {
	UserID uint64  `json:"user_id" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type PaymentRequest struct {
	UserID      uint64  `json:"user_id" validate:"required"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Description string  `json:"description" validate:"required"`
}

type AbsenceRequest struct {
	DriverID    uint64 `json:"driver_id" validate:"required"`
	VehicleType string `json:"vehicle_type" validate:"required,oneof=car_premium car_regular bike"`
}

type WalletResponse struct {
	UserID    uint64  `json:"user_id"`
	Balance   float64 `json:"balance"`
	UpdatedAt string  `json:"updated_at"`
}
