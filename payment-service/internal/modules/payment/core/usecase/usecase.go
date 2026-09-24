package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adyfp24/okejek-go-service/payment-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/dto"
	"github.com/adyfp24/okejek-go-service/payment-service/internal/modules/payment/core/repository"
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

func (uc *UseCase) GetWallet(ctx context.Context, userID uint64) (*dto.WalletResponse, error) {
	wallet, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &dto.WalletResponse{
		UserID:    wallet.UserID,
		Balance:   wallet.Balance,
		UpdatedAt: wallet.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (uc *UseCase) TopUp(ctx context.Context, req dto.TopUpRequest) error {
	desc := "Top-up Saldo Okejek"
	return uc.repo.UpdateBalance(ctx, req.UserID, req.Amount, "credit", desc)
}

func (uc *UseCase) Pay(ctx context.Context, req dto.PaymentRequest) error {
	return uc.repo.UpdateBalance(ctx, req.UserID, req.Amount, "debit", req.Description)
}

// ProcessDriverAbsence memotong biaya absensi harian driver
func (uc *UseCase) ProcessDriverAbsence(ctx context.Context, req dto.AbsenceRequest) error {
	var absenceFee float64
	var vehicleName string

	// Logika bisnis biaya absensi mitra driver Okejek:
	switch req.VehicleType {
	case "car_premium":
		absenceFee = 35000.0 // Biaya absen Car Premium (kemarin sempat bermasalah di monolith)
		vehicleName = "Premium Car"
	case "car_regular":
		absenceFee = 30000.0 // Biaya absen Car Reguler
		vehicleName = "Regular Car"
	case "bike":
		absenceFee = 15000.0 // Biaya absen Motor
		vehicleName = "Oke Bike"
	default:
		return errors.New("tipe kendaraan tidak dikenal untuk absensi")
	}

	desc := fmt.Sprintf("Biaya absensi harian driver tipe %s", vehicleName)
	err := uc.repo.UpdateBalance(ctx, req.DriverID, absenceFee, "debit", desc)
	if err != nil {
		return fmt.Errorf("gagal memotong biaya absensi: %w", err)
	}

	return nil
}
