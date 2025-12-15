package repository

import (
	"biostat/models"
	"context"
	"time"

	"gorm.io/gorm"
)

type OTPRepository interface {
	Create(ctx context.Context, otp *models.OTPMaster) error
	FindActiveOTP(ctx context.Context, req models.ValidateOTPRequest) (*models.OTPMaster, error)
	IncrementAttempt(ctx context.Context, id uint64) error
	MarkUsed(ctx context.Context, id uint64) error
}
type OTPRepositoryImpl struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &OTPRepositoryImpl{db: db}
}

func (r *OTPRepositoryImpl) Create(ctx context.Context, otp *models.OTPMaster) error {
	return r.db.WithContext(ctx).Create(otp).Error
}

func (r *OTPRepositoryImpl) FindActiveOTP(ctx context.Context, req models.ValidateOTPRequest) (*models.OTPMaster, error) {

	var otp models.OTPMaster

	err := r.db.WithContext(ctx).
		Where(`
            identity_type = ?
            AND identity_value = ?
            AND purpose = ?
            AND is_used = false
            AND expires_at > NOW()
        `,
			req.IdentityType,
			req.IdentityValue,
			req.Purpose,
		).
		Order("created_at desc").
		First(&otp).Error

	return &otp, err
}

func (r *OTPRepositoryImpl) IncrementAttempt(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&models.OTPMaster{}).
		Where("id = ?", id).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).
		Error
}

func (r *OTPRepositoryImpl) MarkUsed(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.OTPMaster{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_used": true,
			"used_at": &now,
		}).Error
}
