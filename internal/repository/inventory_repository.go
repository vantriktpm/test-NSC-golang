package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"url-shortener/internal/models"

	"github.com/google/uuid"
)

// InventoryRepository defines the interface for inventory data operations
type InventoryRepository interface {
	// Product operations
	CreateProduct(ctx context.Context, product *models.Product) error
	GetProduct(ctx context.Context, productID string) (*models.Product, error)
	UpdateProduct(ctx context.Context, product *models.Product) error
	UpdateProductInventory(ctx context.Context, productID string, availableStock, reservedStock, version int) error
	GetAllProducts(ctx context.Context) ([]*models.Product, error)

	// Reservation operations
	CreateReservation(ctx context.Context, reservation *models.UserReservation) error
	GetReservationByID(ctx context.Context, reservationID string) (*models.UserReservation, error)
	GetReservationByOrderID(ctx context.Context, orderID uuid.UUID) (*models.UserReservation, error)
	GetReservationByCorrelationID(ctx context.Context, correlationID uuid.UUID) (*models.UserReservation, error)
	UpdateReservationStatus(ctx context.Context, reservationID, status string) error
	GetExpiredReservations(ctx context.Context) ([]*models.UserReservation, error)
	GetUserReservations(ctx context.Context, userID string) ([]*models.UserReservation, error)

	// Event operations
	CreateInventoryEvent(ctx context.Context, event *models.InventoryEvent) error
	GetInventoryEvents(ctx context.Context, productID string, limit int) ([]*models.InventoryEvent, error)
}

type inventoryRepository struct {
	db *sql.DB
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *sql.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

// Product operations

func (r *inventoryRepository) CreateProduct(ctx context.Context, product *models.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, total_stock, available_stock, reserved_stock, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.TotalStock,
		product.AvailableStock,
		product.ReservedStock,
		product.Version,
	).Scan(&product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *inventoryRepository) GetProduct(ctx context.Context, productID string) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, total_stock, available_stock, reserved_stock, version, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	product := &models.Product{}
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.TotalStock,
		&product.AvailableStock,
		&product.ReservedStock,
		&product.Version,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (r *inventoryRepository) UpdateProduct(ctx context.Context, product *models.Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, price = $4, total_stock = $5, 
		    available_stock = $6, reserved_stock = $7, version = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		product.ID,
		product.Name,
		product.Description,
		product.Price,
		product.TotalStock,
		product.AvailableStock,
		product.ReservedStock,
		product.Version,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *inventoryRepository) UpdateProductInventory(ctx context.Context, productID string, availableStock, reservedStock, version int) error {
	query := `
		UPDATE products
		SET available_stock = $2, reserved_stock = $3, version = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, productID, availableStock, reservedStock, version)
	if err != nil {
		return fmt.Errorf("failed to update product inventory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

func (r *inventoryRepository) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	query := `
		SELECT id, name, description, price, total_stock, available_stock, reserved_stock, version, created_at, updated_at
		FROM products
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.TotalStock,
			&product.AvailableStock,
			&product.ReservedStock,
			&product.Version,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// Reservation operations

func (r *inventoryRepository) CreateReservation(ctx context.Context, reservation *models.UserReservation) error {
	query := `
		INSERT INTO user_reservations (id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		reservation.ID,
		reservation.UserID,
		reservation.ProductID,
		reservation.Quantity,
		reservation.ReservedAt,
		reservation.ExpiresAt,
		reservation.Status,
		reservation.CorrelationID,
	)

	if err != nil {
		return fmt.Errorf("failed to create reservation: %w", err)
	}

	return nil
}

func (r *inventoryRepository) GetReservationByID(ctx context.Context, reservationID string) (*models.UserReservation, error) {
	query := `
		SELECT id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id
		FROM user_reservations
		WHERE id = $1
	`

	reservation := &models.UserReservation{}
	err := r.db.QueryRowContext(ctx, query, reservationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ProductID,
		&reservation.Quantity,
		&reservation.ReservedAt,
		&reservation.ExpiresAt,
		&reservation.Status,
		&reservation.CorrelationID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reservation not found")
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return reservation, nil
}

func (r *inventoryRepository) GetReservationByOrderID(ctx context.Context, orderID uuid.UUID) (*models.UserReservation, error) {
	query := `
		SELECT id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id
		FROM user_reservations
		WHERE correlation_id = $1
	`

	reservation := &models.UserReservation{}
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ProductID,
		&reservation.Quantity,
		&reservation.ReservedAt,
		&reservation.ExpiresAt,
		&reservation.Status,
		&reservation.CorrelationID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reservation not found")
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return reservation, nil
}

func (r *inventoryRepository) GetReservationByCorrelationID(ctx context.Context, correlationID uuid.UUID) (*models.UserReservation, error) {
	query := `
		SELECT id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id
		FROM user_reservations
		WHERE correlation_id = $1
	`

	reservation := &models.UserReservation{}
	err := r.db.QueryRowContext(ctx, query, correlationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.ProductID,
		&reservation.Quantity,
		&reservation.ReservedAt,
		&reservation.ExpiresAt,
		&reservation.Status,
		&reservation.CorrelationID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("reservation not found")
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	return reservation, nil
}

func (r *inventoryRepository) UpdateReservationStatus(ctx context.Context, reservationID, status string) error {
	query := `
		UPDATE user_reservations
		SET status = $2
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, reservationID, status)
	if err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reservation not found")
	}

	return nil
}

func (r *inventoryRepository) GetExpiredReservations(ctx context.Context) ([]*models.UserReservation, error) {
	query := `
		SELECT id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id
		FROM user_reservations
		WHERE expires_at < $1 AND status = 'ACTIVE'
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get expired reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*models.UserReservation
	for rows.Next() {
		reservation := &models.UserReservation{}
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ProductID,
			&reservation.Quantity,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&reservation.CorrelationID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

func (r *inventoryRepository) GetUserReservations(ctx context.Context, userID string) ([]*models.UserReservation, error) {
	query := `
		SELECT id, user_id, product_id, quantity, reserved_at, expires_at, status, correlation_id
		FROM user_reservations
		WHERE user_id = $1
		ORDER BY reserved_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user reservations: %w", err)
	}
	defer rows.Close()

	var reservations []*models.UserReservation
	for rows.Next() {
		reservation := &models.UserReservation{}
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ProductID,
			&reservation.Quantity,
			&reservation.ReservedAt,
			&reservation.ExpiresAt,
			&reservation.Status,
			&reservation.CorrelationID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan reservation: %w", err)
		}
		reservations = append(reservations, reservation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reservations: %w", err)
	}

	return reservations, nil
}

// Event operations

func (r *inventoryRepository) CreateInventoryEvent(ctx context.Context, event *models.InventoryEvent) error {
	query := `
		INSERT INTO inventory_events (id, event_id, event_type, product_id, user_id, quantity, correlation_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		uuid.New().String(),
		event.EventID,
		event.EventType,
		event.ProductID,
		event.UserID,
		event.Quantity,
		event.CorrelationID,
		nil, // metadata - would need JSON serialization
	)

	if err != nil {
		return fmt.Errorf("failed to create inventory event: %w", err)
	}

	return nil
}

func (r *inventoryRepository) GetInventoryEvents(ctx context.Context, productID string, limit int) ([]*models.InventoryEvent, error) {
	query := `
		SELECT event_id, event_type, product_id, user_id, quantity, correlation_id, processed_at
		FROM inventory_events
		WHERE product_id = $1
		ORDER BY processed_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, productID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory events: %w", err)
	}
	defer rows.Close()

	var events []*models.InventoryEvent
	for rows.Next() {
		event := &models.InventoryEvent{}
		var processedAt time.Time
		err := rows.Scan(
			&event.EventID,
			&event.EventType,
			&event.ProductID,
			&event.UserID,
			&event.Quantity,
			&event.CorrelationID,
			&processedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		event.Timestamp = processedAt
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}
