package dto

type CreateOrderRequest struct {
	PaymentDelayMS int     `json:"payment_delay_ms" validate:"gte=0,lte=2000"`
	UserID         uint64  `json:"user_id" validate:"required"`
	OriginLat      float64 `json:"origin_lat" validate:"required"`
	OriginLng      float64 `json:"origin_lng" validate:"required"`
	DestinationLat float64 `json:"destination_lat" validate:"required"`
	DestinationLng float64 `json:"destination_lng" validate:"required"`
	Fare           float64 `json:"fare" validate:"required,gt=0"`
}

type OrderResponse struct {
	ID        uint64  `json:"order_id"`
	UserID    uint64  `json:"user_id"`
	DriverID  *uint64 `json:"driver_id,omitempty"`
	Fare      float64 `json:"fare"`
	Status    string  `json:"status"`
	UserName  string  `json:"user_name"`
	CreatedAt string  `json:"created_at"`
}
