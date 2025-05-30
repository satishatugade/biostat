package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderMaster struct {
	OrderID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:order_id" json:"order_id"`
	OrderAddress  string    `gorm:"column:order_address" json:"order_address"`
	OrderNote     string    `gorm:"column:order_note" json:"order_note"`
	VendorID      uint64    `gorm:"column:vendor_id" json:"vendor_id"`
	VendorType    string    `gorm:"type:varchar(255);column:vendor_type" json:"vendor_type"`
	OrderStatus   string    `gorm:"type:varchar(255);default:pending;column:order_status" json:"order_status"`
	TransactionID string    `gorm:"type:varchar(255);column:transaction_id" json:"transaction_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

func (OrderMaster) TableName() string {
	return "tbl_order_master"
}

type OrderItem struct {
	OrderItemID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:order_item_id" json:"order_item_id"`
	ItemType    string    `gorm:"type:varchar(255);column:item_type" json:"item_type"`
	ItemID      int64     `gorm:"column:item_id" json:"item_id"`
	Quantity    int       `gorm:"column:quantity" json:"quantity"`
	CreatedAt   time.Time `gorm:"autoCreateTime;column:created_at" json:"created_at"`
}

func (OrderItem) TableName() string {
	return "tbl_order_item"
}

type OrderOrderItemMapping struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;column:order_id" json:"order_id"`
	OrderItemID uuid.UUID `gorm:"type:uuid;column:order_item_id" json:"order_item_id"`
}

func (OrderOrderItemMapping) TableName() string {
	return "tbl_order_order_item_mapping"
}

type UserOrderMapping struct {
	ID      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();column:id" json:"id"`
	OrderID uuid.UUID `gorm:"type:uuid;column:order_id" json:"order_id"`
	UserID  uint64    `gorm:"column:user_id" json:"user_id"`
}

func (UserOrderMapping) TableName() string {
	return "tbl_user_order_mapping"
}

type OrderItemRequest struct {
	ItemID   int64  `json:"item_id"`
	ItemType string `json:"item_type"`
	Quantity int    `json:"quantity"`
}

type CreateOrderRequest struct {
	OrderAddress string             `json:"order_address"`
	OrderNote    string             `json:"order_note"`
	VendorID     uint64             `json:"vendor_id"`
	VendorType   string             `json:"vendor_type"`
	Items        []OrderItemRequest `json:"items"`
}

type OrderWithItems struct {
	OrderMaster
	Items []OrderItem `json:"items"`
}

type OrderResponse struct {
	OrderID       uuid.UUID          `json:"order_id"`
	OrderAddress  string             `json:"order_address"`
	OrderNote     string             `json:"order_note"`
	VendorDetails VendorDetails      `json:"vendor_details"`
	OrderStatus   string             `json:"order_status"`
	Items         []OrderItemDetails `json:"items"`
	CreatedAt     time.Time          `json:"created_at"`
}

type VendorDetails struct {
	VendorID   uint64 `json:"vendor_id"`
	VendorType string `json:"vendor_type"`
	VendorName string `json:"vendor_name"`
}

type OrderItemDetails struct {
	OrderItemID     uuid.UUID `json:"order_item_id"`
	ItemName        string    `json:"item_name"`
	ItemDescription string    `json:"item_description"`
	ItemType        string    `json:"item_type"`
	Quantity        int       `json:"quantity"`
}
