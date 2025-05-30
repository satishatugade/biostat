package repository

import (
	"biostat/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateOrder(order *models.OrderMaster, tx *gorm.DB) error
	CreateOrderItem(item *models.OrderItem, tx *gorm.DB) error
	CreateOrderItemMapping(mapping *models.OrderOrderItemMapping, tx *gorm.DB) error
	CreateUserOrderMapping(mapping *models.UserOrderMapping, tx *gorm.DB) error

	GetUserOrderMapping(userID uint64) ([]models.UserOrderMapping, error)
	GetOrderByID(orderID uuid.UUID) (*models.OrderMaster, error)
	GetOrderItemMappings(orderID uuid.UUID) ([]models.OrderOrderItemMapping, error)
	GetOrderItemByID(itemID uuid.UUID) (*models.OrderItem, error)
	GetPrescriptionByID(id int64) (*models.PrescriptionDetail, error)
	GetMedicationByID(id int64) (*models.Medication, error)
	GetVendorByID(id uint64) (*models.SystemUser_, error)
}

type OrderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	if db == nil {
		panic("database instance is null")
	}
	return &OrderRepositoryImpl{db: db}
}

func (r *OrderRepositoryImpl) CreateOrder(order *models.OrderMaster, tx *gorm.DB) error {
	return tx.Create(order).Error
}

func (r *OrderRepositoryImpl) CreateOrderItem(item *models.OrderItem, tx *gorm.DB) error {
	return tx.Create(item).Error
}

func (r *OrderRepositoryImpl) CreateOrderItemMapping(mapping *models.OrderOrderItemMapping, tx *gorm.DB) error {
	return tx.Create(mapping).Error
}

func (r *OrderRepositoryImpl) CreateUserOrderMapping(mapping *models.UserOrderMapping, tx *gorm.DB) error {
	return tx.Create(mapping).Error
}

func (r *OrderRepositoryImpl) GetUserOrderMapping(userID uint64) ([]models.UserOrderMapping, error) {
	var userOrderMappings []models.UserOrderMapping
	err := r.db.Where("user_id=?", userID).Find(&userOrderMappings).Error
	return userOrderMappings, err
}

func (r *OrderRepositoryImpl) GetOrderByID(orderID uuid.UUID) (*models.OrderMaster, error) {
	var order models.OrderMaster
	err := r.db.Where("order_id = ?", orderID).First(&order).Error
	return &order, err
}

func (r *OrderRepositoryImpl) GetOrderItemMappings(orderID uuid.UUID) ([]models.OrderOrderItemMapping, error) {
	var mappings []models.OrderOrderItemMapping
	err := r.db.Where("order_id = ?", orderID).Find(&mappings).Error
	return mappings, err
}

func (r *OrderRepositoryImpl) GetOrderItemByID(itemID uuid.UUID) (*models.OrderItem, error) {
	var item models.OrderItem
	err := r.db.Where("order_item_id = ?", itemID).First(&item).Error
	return &item, err
}

func (r *OrderRepositoryImpl) GetPrescriptionByID(id int64) (*models.PrescriptionDetail, error) {
	var pres models.PrescriptionDetail
	err := r.db.Where("prescription_detail_id = ?", id).First(&pres).Error
	return &pres, err
}

func (r *OrderRepositoryImpl) GetMedicationByID(id int64) (*models.Medication, error) {
	var med models.Medication
	err := r.db.Where("medication_id = ?", id).First(&med).Error
	return &med, err
}

func (r *OrderRepositoryImpl) GetVendorByID(id uint64) (*models.SystemUser_, error) {
	var user models.SystemUser_
	err := r.db.Where("user_id=?", id).First(&user).Error
	return &user, err
}
