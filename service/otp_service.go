package service

import (
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"context"
	"errors"
	"time"
)

type OTPService interface {
	GenerateOTPService(req models.GenerateOTPRequest, notifyId string) error
	ValidateOTPService(req models.ValidateOTPRequest) error
}

type OTPServiceImpl struct {
	otpRepository       repository.OTPRepository
	notificationService NotificationService
}

func NewOTPService(otpRepository repository.OTPRepository, notificationService NotificationService) OTPService {
	return &OTPServiceImpl{otpRepository: otpRepository, notificationService: notificationService}
}

func (s *OTPServiceImpl) GenerateOTPService(req models.GenerateOTPRequest, notifyId string) error {
	otp := utils.GenerateOTP(6)

	hash, err := utils.HashOTP(otp)
	if err != nil {
		return err
	}

	record := models.OTPMaster{
		UserID:        req.UserID,
		IdentityType:  req.IdentityType,
		IdentityValue: req.IdentityValue,
		Purpose:       req.Purpose,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		FamilyID:      req.FamilyID,
		OTPHash:       hash,
		Channel:       req.Channel,
		ExpiresAt:     time.Now().Add(5 * time.Minute),
		MaxAttempts:   5,
	}

	if err := s.otpRepository.Create(context.Background(), &record); err != nil {
		return err
	}

	return s.notificationService.SendMemberRemoveOTPMail(notifyId, otp)
}

func (s *OTPServiceImpl) ValidateOTPService(req models.ValidateOTPRequest) error {
	ctx := context.Background()

	otpRecord, err := s.otpRepository.FindActiveOTP(ctx, req)
	if err != nil {
		return errors.New("invalid or expired OTP")
	}

	if otpRecord.AttemptCount >= otpRecord.MaxAttempts {
		return errors.New("maximum OTP attempts exceeded")
	}

	if !utils.CompareOTP(otpRecord.OTPHash, req.OTP) {
		_ = s.otpRepository.IncrementAttempt(ctx, otpRecord.ID)
		return errors.New("invalid OTP")
	}

	return s.otpRepository.MarkUsed(ctx, otpRecord.ID)
}
