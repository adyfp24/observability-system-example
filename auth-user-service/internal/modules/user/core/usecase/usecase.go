package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/app/config"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/dto"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/entity"
	"github.com/adyfp24/okejek-go-service/auth-user-service/internal/modules/user/core/repository"
	"golang.org/x/crypto/bcrypt"
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

func (uc *UseCase) Register(ctx context.Context, req dto.RegisterRequest) error {
	// 1. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashedPassword),
		Role:        req.Role,
	}

	// 2. Simpan user ke database
	err = uc.repo.Create(ctx, user)
	if err != nil {
		return errors.New("nomor handphone sudah terdaftar")
	}

	// 3. Jika rolenya driver, daftarkan juga ke tabel driver
	if req.Role == "driver" {
		driver := &entity.Driver{
			UserID:      user.ID,
			VehicleType: req.VehicleType,
			IsOnline:    false,
		}
		return uc.repo.CreateDriver(ctx, driver)
	}

	return nil
}

func (uc *UseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	// 1. Cari user berdasarkan nomor telepon
	user, err := uc.repo.FindByPhoneNumber(ctx, req.PhoneNumber)
	if err != nil {
		return nil, errors.New("nomor handphone atau password salah")
	}

	// 2. Verifikasi hash password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("nomor handphone atau password salah")
	}

	// 3. Buat token JWT tiruan (Mock Token)
	// Pada implementasi riil, Anda akan menggunakan jwt-go untuk mengodekan user.ID & role
	mockToken := "mock-jwt-token-for-user-" + string(user.Role) + "-" + time.Now().Format("20060102150405")

	return &dto.LoginResponse{
		Token:  mockToken,
		UserID: user.ID,
		Role:   user.Role,
	}, nil
}

func (uc *UseCase) GetProfile(ctx context.Context, id uint64) (*dto.UserProfile, error) {
	user, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	return &dto.UserProfile{
		ID:          user.ID,
		Name:        user.Name,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
	}, nil
}
