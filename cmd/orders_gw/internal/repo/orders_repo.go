package repo

import (
	"context"

	"go-gw-test/cmd/orders_gw/internal/types"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// OrdersRepo defines persistence operations for orders_gw.
type OrdersRepo interface {
	ListOrders(ctx context.Context) ([]types.OrderRecord, error)
	FindOrderByID(ctx context.Context, orderID int64) (types.OrderRecord, error)
	ListOrderItems(ctx context.Context, orderID int64) ([]types.OrderItem, error)
	SeedIfEmpty(ctx context.Context) error
}

// OrdersRepoImpl implements OrdersRepo using GORM.
type OrdersRepoImpl struct {
	db *gorm.DB
}

// NewOrdersRepo constructs an OrdersRepo implementation.
func NewOrdersRepo(db *gorm.DB) *OrdersRepoImpl {
	return &OrdersRepoImpl{
		db: db,
	}
}

// ListOrders returns all orders.
func (r *OrdersRepoImpl) ListOrders(ctx context.Context) ([]types.OrderRecord, error) {
	var orders []types.OrderRecord
	err := r.db.WithContext(ctx).Find(&orders).Error
	if err != nil {
		zap.L().Error("list orders", zap.Error(err))
		return nil, err
	}

	return orders, nil
}

// FindOrderByID returns an order by ID.
func (r *OrdersRepoImpl) FindOrderByID(ctx context.Context, orderID int64) (types.OrderRecord, error) {
	var order types.OrderRecord
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&order).Error
	if err != nil {
		zap.L().Error("find order", zap.Error(err))
		return types.OrderRecord{}, err
	}

	return order, nil
}

// ListOrderItems returns items for a given order.
func (r *OrdersRepoImpl) ListOrderItems(ctx context.Context, orderID int64) ([]types.OrderItem, error) {
	var items []types.OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error
	if err != nil {
		zap.L().Error("list order items", zap.Error(err))
		return nil, err
	}

	return items, nil
}

// SeedIfEmpty inserts sample orders when no records exist.
func (r *OrdersRepoImpl) SeedIfEmpty(ctx context.Context) error {
	var count int64
	err := r.db.WithContext(ctx).Model(&types.OrderRecord{}).Count(&count).Error
	if err != nil {
		zap.L().Error("count orders", zap.Error(err))
		return err
	}

	if count == 0 {
		seedOrders := []types.OrderRecord{
			{ID: 1, UserID: 1, Status: "processing"},
			{ID: 2, UserID: 2, Status: "shipped"},
			{ID: 3, UserID: 1, Status: "delivered"},
		}

		err = r.db.WithContext(ctx).Create(&seedOrders).Error
		if err != nil {
			zap.L().Error("seed orders", zap.Error(err))
			return err
		}
	}

	var itemCount int64
	err = r.db.WithContext(ctx).Model(&types.OrderItem{}).Count(&itemCount).Error
	if err != nil {
		zap.L().Error("count order items", zap.Error(err))
		return err
	}

	if itemCount > 0 {
		return nil
	}

	seedItems := []types.OrderItem{
		{ID: 1, OrderID: 1, SKU: "starter-kit", Name: "Starter Kit", Quantity: 1, UnitPrice: 49.99},
		{ID: 2, OrderID: 2, SKU: "premium-plan", Name: "Premium Plan", Quantity: 1, UnitPrice: 199.0},
		{ID: 3, OrderID: 3, SKU: "accessory-pack", Name: "Accessory Pack", Quantity: 2, UnitPrice: 19.5},
		{ID: 4, OrderID: 3, SKU: "warranty", Name: "Extended Warranty", Quantity: 1, UnitPrice: 29.0},
	}

	err = r.db.WithContext(ctx).Create(&seedItems).Error
	if err != nil {
		zap.L().Error("seed order items", zap.Error(err))
		return err
	}

	return nil
}
