package service

import (
	"biostat/constant"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"fmt"
	"log"
	"time"
)

type SubscriptionService interface {
	SubscribePlan(req models.SubscribeFamilyRequest, userId uint64, userInfo models.SystemUser_) (map[string]interface{}, bool, error)
	GetActiveSubscriptionPlanByMemberId(userId uint64) (*models.SubscriptionMaster, bool, constant.SubscriptionStatus, string, error)
	SubscribeDefaultPlan(userId uint64, roleName string, lastName string) error
	UpdateSubscriptionStatus(enabled bool, updatedBy string) error
	GetSubscriptionShowStatus() (bool, error)
	FetchSubscriptionPlanService() ([]models.SubscriptionMaster, error)
	ValidateFamilyMemberLimit(memberId uint64, roleName string) error
}

type SubscriptionServiceImpl struct {
	subscriptionRepo repository.SubscriptionRepository
	roleRepo         repository.RoleRepository
}

// FetchSubscriptionPlanService implements SubscriptionService.
func (s *SubscriptionServiceImpl) FetchSubscriptionPlanService() ([]models.SubscriptionMaster, error) {
	return s.subscriptionRepo.GetSubscriptionWithServices()
}

func NewSubscriptionService(subscriptionRepo repository.SubscriptionRepository, roleRepo repository.RoleRepository) SubscriptionService {
	return &SubscriptionServiceImpl{subscriptionRepo: subscriptionRepo, roleRepo: roleRepo}
}

func (s *SubscriptionServiceImpl) GetActiveSubscriptionPlanByMemberId(memberId uint64) (*models.SubscriptionMaster, bool, constant.SubscriptionStatus, string, error) {
	return s.subscriptionRepo.CheckActiveSubscriptionPlanByMemberId(memberId)
}

func (s *SubscriptionServiceImpl) SubscribePlan(req models.SubscribeFamilyRequest, userId uint64, userInfo models.SystemUser_) (map[string]interface{}, bool, error) {
	existingPlan, planExist, _, _, existErr := s.subscriptionRepo.CheckActiveSubscriptionPlanByMemberId(userId)
	log.Println("CheckActiveSubscriptionPlanByMemberId : ", planExist, existingPlan)
	if existErr != nil {
		return nil, false, fmt.Errorf("error while Fetching plan details")
	}
	if planExist && req.RenewalType != "UPGRADE" {
		return map[string]interface{}{
			"family_id":               existingPlan.FamilyId,
			"member_id":               existingPlan.MemberId,
			"plan_name":               existingPlan.PlanName,
			"subscription_start_date": existingPlan.SubscriptionStartDate,
			"subscription_end_date":   existingPlan.SubscriptionEndDate,
			"is_auto_renew":           existingPlan.IsAutoRenew,
			"duration":                existingPlan.Duration,
			"price":                   existingPlan.Price,
			// "updated_by":              existingPlan.MemberId,
		}, true, nil
	}

	plan, err := s.subscriptionRepo.GetSubsciptionMasterPlanBySubscriptionId(req.SubscriptionId)
	if err != nil {
		return nil, false, fmt.Errorf("subscription plan not found")
	}
	family, err := s.subscriptionRepo.GetPatientFamilyGroupSubscriptionInfoByFamilyId(req.FamilyId)
	if err != nil {
		family, err = s.GetOrCreateFamilyByMemberID(userId, userInfo.LastName)
		if err != nil {
			return nil, false, fmt.Errorf("failed to get or create family: %w", err)
		}
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, plan.Duration)

	family.MemberId = userId
	family.CurrentSubscriptionId = &plan.SubscriptionId
	family.SubscriptionStartDate = &now
	family.SubscriptionEndDate = &endDate
	family.IsAutoRenew = req.IsAutoRenew
	family.LastRenewedAt = &now
	family.LastRenewalType = req.RenewalType
	family.LastRenewedBy = userId
	family.IsActive = true

	if _, err := s.subscriptionRepo.UpdateFamilySubscription(family); err != nil {
		return nil, false, fmt.Errorf("failed to update subscription")
	}
	if err := s.roleRepo.UpdateFamilyIdSystemRoleMapping(family.FamilyId, userId); err != nil {
		return nil, false, fmt.Errorf("failed to  UpdateFamilyIdSystemRoleMapping: %w", err)
	}

	return map[string]interface{}{
		"family_id":               family.FamilyId,
		"member_id":               family.MemberId,
		"plan_name":               plan.PlanName,
		"subscription_start_date": utils.FormatDateTime(family.SubscriptionStartDate),
		"subscription_end_date":   utils.FormatDateTime(family.SubscriptionEndDate),
		"is_auto_renew":           family.IsAutoRenew,
		"duration":                plan.Duration,
		"price":                   plan.Price,
		// "updated_by":              userId,
	}, false, nil
}

func (s *SubscriptionServiceImpl) SubscribeDefaultPlan(userId uint64, roleName, lastName string) error {
	if roleName != string(constant.Patient) {
		return fmt.Errorf("user role is not patient ")
	}
	enabled, err := s.GetSubscriptionShowStatus()
	if err != nil {
		return fmt.Errorf("failed to fetch subscription status: %w", err)
	}
	if !enabled {
		return fmt.Errorf("user role is not patient %t", enabled)
	}
	plan, err := s.subscriptionRepo.GetSubsciptionMasterPlanBySubscriptionId(1)
	if err != nil {
		return fmt.Errorf("default subscription plan not found")
	}

	family, err := s.subscriptionRepo.GetFamilyGroupByMemberID(userId)
	if err != nil {
		family, err = s.GetOrCreateFamilyByMemberID(userId, lastName)
		if err != nil {
			return fmt.Errorf("failed to get or create family: %w", err)
		}
	}

	if family.CurrentSubscriptionId != nil && *family.CurrentSubscriptionId == plan.SubscriptionId {
		return nil
	}

	now := time.Now()
	endDate := now.AddDate(0, 0, plan.Duration)

	family.MemberId = userId
	family.CurrentSubscriptionId = &plan.SubscriptionId
	family.SubscriptionStartDate = &now
	family.SubscriptionEndDate = &endDate
	family.IsAutoRenew = false
	family.LastRenewedAt = &now
	family.LastRenewalType = "DEFAULT"
	family.LastRenewedBy = userId
	family.IsActive = true

	familyId, err := s.subscriptionRepo.UpdateFamilySubscription(family)
	if err != nil {
		return fmt.Errorf("failed to subscribe user to default plan: %w", err)
	}
	if err := s.roleRepo.UpdateFamilyIdSystemRoleMapping(familyId, userId); err != nil {
		return fmt.Errorf("failed to  UpdateFamilyIdSystemRoleMapping: %w", err)
	}

	return nil
}

func (s *SubscriptionServiceImpl) GetOrCreateFamilyByMemberID(userId uint64, lastName string) (*models.PatientFamilyGroup, error) {
	family, err := s.subscriptionRepo.GetFamilyGroupByMemberID(userId)
	if err == nil && family != nil {
		return family, nil
	}

	newFamily := &models.PatientFamilyGroup{
		FamilyName: fmt.Sprintf("%s-%d", lastName, time.Now().Unix()),
		MemberId:   userId,
		CreatedAt:  time.Now(),
	}

	if err := s.subscriptionRepo.CreateFamily(newFamily); err != nil {
		return nil, fmt.Errorf("failed to create family group: %w", err)
	}

	return newFamily, nil
}

func (s *SubscriptionServiceImpl) GetSubscriptionShowStatus() (bool, error) {
	return s.subscriptionRepo.GetSubscriptionShowStatus()
}

func (s *SubscriptionServiceImpl) UpdateSubscriptionStatus(enabled bool, updatedBy string) error {
	return s.subscriptionRepo.UpdateSubscriptionStatus(enabled, updatedBy)
}

func (s *SubscriptionServiceImpl) ValidateFamilyMemberLimit(memberId uint64, roleName string) error {
	if roleName != string(constant.Relative) {
		return nil
	}
	enabled, err := s.GetSubscriptionShowStatus()
	if err != nil {
		return fmt.Errorf("failed to fetch subscription status: %w", err)
	}
	// If subscription model disabled → always allow
	if !enabled {
		return nil
	}

	if memberId == 0 {
		return fmt.Errorf("family ID required for subscription check")
	}

	family, err := s.subscriptionRepo.GetFamilyGroupByMemberID(memberId)
	if err != nil {
		return fmt.Errorf("cannot fetch family for member: %w", err)
	}

	// 3. If no active subscription → block
	if family.CurrentSubscriptionId == nil {
		return fmt.Errorf("no active subscription for this family. Please subscribe to a plan")
	}

	plan, err := s.subscriptionRepo.GetSubsciptionMasterPlanBySubscriptionId(*family.CurrentSubscriptionId)
	if err != nil {
		return fmt.Errorf("subscription plan not found")
	}

	memberCount, err := s.roleRepo.GetCountFamilyMember(memberId, family.FamilyId)
	if err != nil {
		return fmt.Errorf("failed to count family members")
	}

	// Compare with plan limit
	if memberCount >= plan.MaxMember {
		return fmt.Errorf("subscription limit reached (%d/%d). Please upgrade your plan", memberCount, plan.MaxMember)
	}

	return nil
}
