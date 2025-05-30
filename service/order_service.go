package service

import (
	"biostat/database"
	"biostat/models"
	"biostat/repository"
	"biostat/utils"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(request *models.CreateOrderRequest, userID uint64) (*models.OrderMaster, error)
	GetOrdersByUserID(userID uint64) ([]models.OrderResponse, error)
}

type OrderServiceImpl struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &OrderServiceImpl{orderRepo: orderRepo}
}

func (s *OrderServiceImpl) GetOrdersByUserID(userID uint64) ([]models.OrderResponse, error) {
	mappings, err := s.orderRepo.GetUserOrderMapping(userID)
	if err != nil {
		return nil, err
	}
	var responses []models.OrderResponse
	for _, m := range mappings {
		order, err := s.orderRepo.GetOrderByID(m.OrderID)
		if err != nil {
			continue
		}
		vendorName := ""
		if order.VendorType == "pharmacist" {
			vendor, _ := s.orderRepo.GetVendorByID(order.VendorID)
			if vendor != nil {
				mappedUser := utils.MapUserToRoleSchema(*vendor, "pharmacist")
				pharmacistDetails, ok := mappedUser.(models.Pharmacist)
				if ok {
					vendorName = pharmacistDetails.PharmacyName
				}
			}
		}

		vendorDetails := models.VendorDetails{
			VendorID:   order.VendorID,
			VendorType: order.VendorType,
			VendorName: vendorName,
		}

		var items []models.OrderItemDetails
		orderItemMappings, _ := s.orderRepo.GetOrderItemMappings(order.OrderID)
		for _, itemMapping := range orderItemMappings {
			item, err := s.orderRepo.GetOrderItemByID(itemMapping.OrderItemID)
			if err != nil {
				continue
			}
			itemName, itemDesc := "", ""
			switch item.ItemType {
			case "medication":
				med, _ := s.orderRepo.GetMedicationByID(item.ItemID)
				if med != nil {
					itemName = med.MedicationName
					itemDesc = med.Description
				}
			case "prescription":
				prescription, _ := s.orderRepo.GetPrescriptionByID(item.ItemID)

				if prescription != nil {
					itemName = fmt.Sprintf("%s (%.0f %s) - %s", prescription.MedicineName, prescription.UnitValue, prescription.UnitType, prescription.PrescriptionType)
					itemDesc = prescription.Instruction
				}
			}

			items = append(items, models.OrderItemDetails{
				OrderItemID:     item.OrderItemID,
				ItemType:        item.ItemType,
				ItemName:        itemName,
				ItemDescription: itemDesc,
				Quantity:        item.Quantity,
			})
		}
		responses = append(responses, models.OrderResponse{
			OrderID:       order.OrderID,
			OrderAddress:  order.OrderAddress,
			OrderNote:     order.OrderNote,
			VendorDetails: vendorDetails,
			OrderStatus:   order.OrderStatus,
			Items:         items,
			CreatedAt:     order.CreatedAt,
		})

	}

	return responses, nil

}

func (s *OrderServiceImpl) CreateOrder(req *models.CreateOrderRequest, userID uint64) (*models.OrderMaster, error) {
	if len(req.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("Recovered in CreateOrder:", r)
			return
		}
	}()
	order := &models.OrderMaster{
		OrderID:      uuid.New(),
		OrderAddress: req.OrderAddress,
		OrderNote:    req.OrderNote,
		VendorID:     req.VendorID,
		VendorType:   req.VendorType,
		OrderStatus:  "pending",
	}
	if err := s.orderRepo.CreateOrder(order, tx); err != nil {
		tx.Rollback()
		log.Println("@CreateOrder->CreateOrder err:", err)
		return nil, err
	}
	userMapping := &models.UserOrderMapping{
		ID:      uuid.New(),
		OrderID: order.OrderID,
		UserID:  userID,
	}
	if err := s.orderRepo.CreateUserOrderMapping(userMapping, tx); err != nil {
		tx.Rollback()
		log.Println("@CreateOrder->CreateUserOrderMapping err:", err)
		return nil, err
	}
	for _, item := range req.Items {
		orderItem := &models.OrderItem{
			OrderItemID: uuid.New(),
			ItemType:    item.ItemType,
			ItemID:      item.ItemID,
			Quantity:    item.Quantity,
		}
		if err := s.orderRepo.CreateOrderItem(orderItem, tx); err != nil {
			tx.Rollback()
			log.Println("@CreateOrder->CreateOrderItem err:", err)
			return nil, err
		}
		mapping := &models.OrderOrderItemMapping{
			ID:          uuid.New(),
			OrderID:     order.OrderID,
			OrderItemID: orderItem.OrderItemID,
		}
		if err := s.orderRepo.CreateOrderItemMapping(mapping, tx); err != nil {
			tx.Rollback()
			log.Println("@CreateOrder->CreateOrderItemMapping err:", err)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Println("@CreateOrder->Commit err:", err)
		return nil, err
	}

	return order, nil
}
