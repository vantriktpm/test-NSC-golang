-- Inventory Management Database Schema
-- This schema supports Kafka-based inventory management system

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    total_stock INTEGER NOT NULL DEFAULT 0 CHECK (total_stock >= 0),
    available_stock INTEGER NOT NULL DEFAULT 0 CHECK (available_stock >= 0),
    reserved_stock INTEGER NOT NULL DEFAULT 0 CHECK (reserved_stock >= 0),
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT check_stock_consistency CHECK (total_stock = available_stock + reserved_stock)
);

-- User reservations table
CREATE TABLE IF NOT EXISTS user_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    reserved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'CONFIRMED', 'RELEASED', 'EXPIRED')),
    correlation_id UUID NOT NULL,
    
    -- Indexes
    INDEX idx_user_reservations_user_id (user_id),
    INDEX idx_user_reservations_product_id (product_id),
    INDEX idx_user_reservations_expires_at (expires_at),
    INDEX idx_user_reservations_correlation_id (correlation_id),
    INDEX idx_user_reservations_status (status)
);

-- Inventory events table (for audit trail and debugging)
CREATE TABLE IF NOT EXISTS inventory_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL CHECK (event_type IN ('INVENTORY_CHECK', 'INVENTORY_RESERVE', 'INVENTORY_CONFIRM', 'INVENTORY_RELEASE', 'INVENTORY_RESTOCK')),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id UUID,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    correlation_id UUID,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB,
    
    -- Indexes
    INDEX idx_inventory_events_product_id (product_id),
    INDEX idx_inventory_events_event_type (event_type),
    INDEX idx_inventory_events_processed_at (processed_at),
    INDEX idx_inventory_events_correlation_id (correlation_id)
);

-- Orders table (for completed purchases)
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    status VARCHAR(20) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'CONFIRMED', 'CANCELLED', 'COMPLETED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    correlation_id UUID NOT NULL,
    
    -- Indexes
    INDEX idx_orders_user_id (user_id),
    INDEX idx_orders_product_id (product_id),
    INDEX idx_orders_status (status),
    INDEX idx_orders_created_at (created_at),
    INDEX idx_orders_correlation_id (correlation_id)
);

-- Inventory snapshots table (for historical tracking)
CREATE TABLE IF NOT EXISTS inventory_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    available_stock INTEGER NOT NULL,
    reserved_stock INTEGER NOT NULL,
    total_stock INTEGER NOT NULL,
    version INTEGER NOT NULL,
    snapshot_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    INDEX idx_inventory_snapshots_product_id (product_id),
    INDEX idx_inventory_snapshots_snapshot_at (snapshot_at)
);

-- Triggers for automatic timestamp updates
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_products_updated_at BEFORE UPDATE ON products
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to create inventory snapshot
CREATE OR REPLACE FUNCTION create_inventory_snapshot(product_uuid UUID)
RETURNS VOID AS $$
BEGIN
    INSERT INTO inventory_snapshots (product_id, available_stock, reserved_stock, total_stock, version)
    SELECT id, available_stock, reserved_stock, total_stock, version
    FROM products
    WHERE id = product_uuid;
END;
$$ LANGUAGE plpgsql;

-- Function to cleanup expired reservations
CREATE OR REPLACE FUNCTION cleanup_expired_reservations()
RETURNS INTEGER AS $$
DECLARE
    expired_count INTEGER;
BEGIN
    UPDATE user_reservations
    SET status = 'EXPIRED'
    WHERE expires_at < CURRENT_TIMESTAMP AND status = 'ACTIVE';
    
    GET DIAGNOSTICS expired_count = ROW_COUNT;
    RETURN expired_count;
END;
$$ LANGUAGE plpgsql;

-- Views for common queries

-- Product availability view
CREATE OR REPLACE VIEW product_availability AS
SELECT 
    p.id,
    p.name,
    p.price,
    p.available_stock,
    p.reserved_stock,
    p.total_stock,
    CASE 
        WHEN p.available_stock > 0 THEN 'IN_STOCK'
        WHEN p.reserved_stock > 0 THEN 'RESERVED'
        ELSE 'OUT_OF_STOCK'
    END as stock_status,
    p.updated_at
FROM products p;

-- User reservation summary view
CREATE OR REPLACE VIEW user_reservation_summary AS
SELECT 
    ur.user_id,
    COUNT(*) as total_reservations,
    SUM(ur.quantity) as total_quantity,
    COUNT(CASE WHEN ur.status = 'ACTIVE' THEN 1 END) as active_reservations,
    COUNT(CASE WHEN ur.status = 'CONFIRMED' THEN 1 END) as confirmed_reservations,
    COUNT(CASE WHEN ur.status = 'RELEASED' THEN 1 END) as released_reservations
FROM user_reservations ur
GROUP BY ur.user_id;

-- Inventory metrics view
CREATE OR REPLACE VIEW inventory_metrics AS
SELECT 
    COUNT(*) as total_products,
    SUM(total_stock) as total_stock_value,
    SUM(available_stock) as total_available_stock,
    SUM(reserved_stock) as total_reserved_stock,
    COUNT(CASE WHEN available_stock = 0 THEN 1 END) as out_of_stock_products,
    COUNT(CASE WHEN available_stock < 10 THEN 1 END) as low_stock_products,
    AVG(available_stock) as avg_available_stock
FROM products;

-- Sample data for testing
INSERT INTO products (id, name, description, price, total_stock, available_stock, reserved_stock) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'iPhone 15 Pro', 'Latest iPhone with advanced features', 999.99, 100, 100, 0),
('550e8400-e29b-41d4-a716-446655440002', 'Samsung Galaxy S24', 'Premium Android smartphone', 899.99, 50, 50, 0),
('550e8400-e29b-41d4-a716-446655440003', 'MacBook Pro M3', 'High-performance laptop', 1999.99, 25, 25, 0),
('550e8400-e29b-41d4-a716-446655440004', 'AirPods Pro', 'Wireless earbuds with noise cancellation', 249.99, 200, 200, 0),
('550e8400-e29b-41d4-a716-446655440005', 'iPad Air', 'Tablet for productivity and entertainment', 599.99, 75, 75, 0)
ON CONFLICT (id) DO NOTHING;
