package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/adyfp24/okejek-go-service/order-service/internal/app/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"net/http"
	"os"
	"time"

	"github.com/adyfp24/okejek-go-service/order-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/dto"
	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/entity"
	"github.com/adyfp24/okejek-go-service/order-service/internal/modules/order/core/repository"
)

type UseCase struct {
	repo      *repository.Repository
	validator *config.Validator
	cfg       *config.Config
}

func NewUseCase(repo *repository.Repository, validator *config.Validator, cfg *config.Config) *UseCase {
	return &UseCase{
		repo:      repo,
		validator: validator,
		cfg:       cfg,
	}
}

// fetchUserProfileFromUserService memanggil HTTP REST ke Auth/User Service
func (uc *UseCase) fetchUserProfileFromUserService(ctx context.Context, userID uint64) (string, string, error) {
	// Menghubungi auth_user_service yang berjalan di port 3003 (sesuai .env port)
	url := fmt.Sprintf("%s/api/v1/users/%d", uc.cfg.Get("USER_SERVICE_URL", "http://localhost:3003"), userID)

	client := &http.Client{Timeout: 5 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("gagal mendapatkan profil dari user service")
	}

	var result struct {
		Name        string `json:"name"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}

	return result.Name, result.PhoneNumber, nil
}

func (uc *UseCase) CreateOrder(ctx context.Context, req dto.CreateOrderRequest) (*entity.Order, error) {
	// 1. Ambil data user dari User Service (Shim Layer / Integration)
	nameSnapshot, phoneSnapshot, err := uc.fetchUserProfileFromUserService(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user lookup failed: %w", err)
	}

	order := &entity.Order{
		UserID:            req.UserID,
		OriginLat:         req.OriginLat,
		OriginLng:         req.OriginLng,
		DestinationLat:    req.DestinationLat,
		DestinationLng:    req.DestinationLng,
		Fare:              req.Fare,
		Status:            "created",
		UserNameSnapshot:  nameSnapshot,
		UserPhoneSnapshot: phoneSnapshot,
	}

	// 2. Simpan order ke local database
	err = uc.repo.Create(ctx, order)
	if err != nil {
		return nil, err
	}

	// Checkout uses a synchronous payment call so the audience can inspect one complete trace.
	payload, _ := json.Marshal(map[string]interface{}{"user_id": req.UserID, "amount": req.Fare, "description": fmt.Sprintf("Order #%d", order.ID)})
	request, err := http.NewRequestWithContext(ctx, "POST", uc.cfg.Get("PAYMENT_SERVICE_URL", "http://localhost:3002")+"/api/v1/wallets/pay", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	if os.Getenv("DEMO_MODE") == "true" && req.PaymentDelayMS > 0 {
		request.Header.Set("X-Demo-Delay-Ms", fmt.Sprint(req.PaymentDelayMS))
	}
	client := &http.Client{Timeout: 5 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)}
	response, err := client.Do(request)
	if err != nil {
		_ = uc.repo.UpdateStatus(ctx, order.ID, "payment_unknown")
		telemetry.Event(ctx, "payment_unknown", map[string]interface{}{"order_id": order.ID})
		return nil, fmt.Errorf("payment outcome unknown; do not retry automatically: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		if err := uc.repo.UpdateStatus(ctx, order.ID, "payment_failed"); err != nil {
			return nil, err
		}
		telemetry.Event(ctx, "payment_failed", map[string]interface{}{"order_id": order.ID})
		return nil, fmt.Errorf("payment rejected for order %d", order.ID)
	}
	if err := uc.repo.UpdateStatus(ctx, order.ID, "paid"); err != nil {
		return nil, err
	}
	order.Status = "paid"
	telemetry.Event(ctx, "order_paid", map[string]interface{}{"order_id": order.ID, "user_id": order.UserID})

	return order, nil
}

func (uc *UseCase) GetOrder(ctx context.Context, id uint64) (*dto.OrderResponse, error) {
	order, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("order tidak ditemukan")
	}

	return &dto.OrderResponse{
		ID:        order.ID,
		UserID:    order.UserID,
		DriverID:  order.DriverID,
		Fare:      order.Fare,
		Status:    order.Status,
		UserName:  order.UserNameSnapshot,
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *UseCase) UpdateOrderStatus(ctx context.Context, id uint64, status string) error {
	return uc.repo.UpdateStatus(ctx, id, status)
}
