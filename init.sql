-- Create ENUMs
CREATE TYPE order_status AS ENUM (
    'pending',
    'preparing',
    'ready',
    'delivered',
    'cancelled'
);

CREATE TYPE unit_type AS ENUM (
    'grams',
    'milliliters',
    'pieces',
    'units'
);

CREATE TYPE transaction_type AS ENUM (
    'addition',
    'deduction',
    'adjustment',
    'waste'
);

CREATE TYPE item_size AS ENUM (
    'small',
    'medium',
    'large'
);

-- Create Tables
CREATE TABLE menu_items (
    menu_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price > 0),
    categories TEXT[] NOT NULL DEFAULT '{}',
    allergens TEXT[] NOT NULL DEFAULT '{}',
    size item_size NOT NULL DEFAULT 'medium',
    customization_options JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inventory (
    ingredient_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 0,
    unit unit_type NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL CHECK (unit_price >= 0),
    reorder_point INTEGER NOT NULL CHECK (reorder_point >= 0),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_item_ingredients (
    menu_item_ingr_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES inventory(ingredient_id),
    quantity DECIMAL(10,2) NOT NULL CHECK (quantity > 0),
    unit unit_type NOT NULL,
    UNIQUE(menu_item_id, ingredient_id)
);

CREATE TABLE orders (
    order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_name VARCHAR(255) NOT NULL,
    special_instructions JSONB,
    total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
    status order_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    order_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_at_time DECIMAL(10,2) NOT NULL CHECK (price_at_time >= 0),
    customizations JSONB
    -- Removed UNIQUE constraint to allow multiple of the same item in an order
);

CREATE TABLE order_status_history (
    order_status_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    old_status order_status NOT NULL,
    new_status order_status NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    change_reason TEXT NOT NULL
);

CREATE TABLE price_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    old_price DECIMAL(10,2) NOT NULL CHECK (old_price >= 0),
    new_price DECIMAL(10,2) NOT NULL CHECK (new_price >= 0),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    change_reason TEXT NOT NULL
);

CREATE TABLE inventory_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ingredient_id UUID NOT NULL REFERENCES inventory(ingredient_id),
    quantity_change DECIMAL(10,2) NOT NULL,
    transaction_type transaction_type NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create Indexes
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_menu_items_price ON menu_items(price);
CREATE INDEX idx_menu_items_categories ON menu_items USING GIN(categories);
CREATE INDEX idx_inventory_quantity ON inventory(quantity);
CREATE INDEX idx_order_items_order_id ON order_items(order_id);

-- Full Text Search Indexes
CREATE INDEX idx_menu_items_search ON menu_items 
    USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));
CREATE INDEX idx_orders_customer_search ON orders 
    USING GIN(to_tsvector('english', customer_name || ' ' || COALESCE(special_instructions::text, '')));

-- Composite Indexes
CREATE INDEX idx_order_status_date ON orders(status, created_at);
CREATE INDEX idx_inventory_stock_price ON inventory(quantity, unit_price);

-- Trigger for UpdatedAt
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for tables with updated_at columns
CREATE TRIGGER update_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_menu_items_updated_at
    BEFORE UPDATE ON menu_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Insert Mock Data
-- Menu Items
INSERT INTO menu_items (name, description, price, categories, allergens, size, customization_options) VALUES
    ('Espresso', 'Strong Italian coffee', 3.50, ARRAY['coffee', 'hot'], ARRAY['caffeine'], 'small', '{"extras": ["extra_shot", "hot_water"]}'),
    ('Cappuccino', 'Espresso with steamed milk', 4.50, ARRAY['coffee', 'hot', 'milk'], ARRAY['caffeine', 'lactose'], 'medium', '{"milk_options": ["whole", "skim", "oat"]}'),
    ('Latte', 'Mild coffee with lots of milk', 4.00, ARRAY['coffee', 'hot', 'milk'], ARRAY['caffeine', 'lactose'], 'medium', '{"milk_options": ["whole", "skim", "oat"], "flavors": ["vanilla", "caramel"]}'),
    ('Croissant', 'Buttery French pastry', 3.00, ARRAY['pastry', 'breakfast'], ARRAY['gluten', 'dairy'], 'medium', '{"warming": true}'),
    ('Chocolate Muffin', 'Rich chocolate muffin', 3.50, ARRAY['pastry', 'dessert'], ARRAY['gluten', 'dairy', 'eggs'], 'medium', '{"warming": true}'),
    ('Green Tea', 'Japanese green tea', 3.00, ARRAY['tea', 'hot'], ARRAY[]::TEXT[], 'medium', '{"strength": ["light", "medium", "strong"]}'),
    ('Iced Coffee', 'Cold brew coffee', 4.00, ARRAY['coffee', 'cold'], ARRAY['caffeine'], 'large', '{"ice": ["normal", "light", "extra"], "sweetener": true}'),
    ('Breakfast Sandwich', 'Egg and cheese sandwich', 6.00, ARRAY['sandwich', 'breakfast'], ARRAY['gluten', 'dairy', 'eggs'], 'medium', '{"bread": ["croissant", "bagel", "english_muffin"]}'),
    ('Cheesecake', 'New York style cheesecake', 5.50, ARRAY['dessert'], ARRAY['gluten', 'dairy', 'eggs'], 'medium', '{"toppings": ["strawberry", "chocolate", "caramel"]}'),
    ('Smoothie', 'Fresh fruit smoothie', 5.00, ARRAY['beverages', 'cold'], ARRAY[]::TEXT[], 'large', '{"fruits": ["strawberry", "banana", "mango"], "extras": ["protein", "spinach"]}');

-- Inventory Items
INSERT INTO inventory (name, quantity, unit, unit_price, reorder_point) VALUES
    ('Coffee Beans', 10000, 'grams', 0.04, 2000),
    ('Whole Milk', 20000, 'milliliters', 0.002, 5000),
    ('Sugar', 5000, 'grams', 0.002, 1000),
    ('Chocolate Powder', 2000, 'grams', 0.05, 500),
    ('Green Tea Leaves', 1000, 'grams', 0.08, 200),
    ('Croissant Dough', 100, 'pieces', 1.00, 20),
    ('Muffin Mix', 5000, 'grams', 0.03, 1000),
    ('Eggs', 200, 'pieces', 0.25, 50),
    ('Cheese', 3000, 'grams', 0.05, 500),
    ('English Muffins', 100, 'pieces', 0.50, 20),
    ('Strawberries', 2000, 'grams', 0.02, 500),
    ('Bananas', 5000, 'grams', 0.01, 1000),
    ('Whipped Cream', 2000, 'grams', 0.03, 500),
    ('Caramel Syrup', 2000, 'milliliters', 0.02, 500),
    ('Vanilla Syrup', 2000, 'milliliters', 0.02, 500),
    ('Ice', 10000, 'grams', 0.001, 2000),
    ('Paper Cups', 500, 'pieces', 0.10, 100),
    ('Napkins', 1000, 'pieces', 0.02, 200),
    ('To-Go Bags', 300, 'pieces', 0.15, 50),
    ('Straws', 800, 'pieces', 0.01, 200);

-- Menu Item Ingredients (Recipe relationships)
INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, quantity, unit) 
SELECT 
    m.menu_item_id,
    i.ingredient_id,
    CASE 
        WHEN m.name = 'Espresso' THEN 18
        WHEN m.name = 'Cappuccino' THEN 18
        WHEN m.name = 'Latte' THEN 14
    END,
    CASE 
        WHEN i.name = 'Coffee Beans' THEN 'grams'::unit_type
        WHEN i.name = 'Whole Milk' THEN 'milliliters'::unit_type
        ELSE 'grams'::unit_type
    END
FROM menu_items m
CROSS JOIN inventory i
WHERE 
    (m.name = 'Espresso' AND i.name = 'Coffee Beans')
    OR (m.name = 'Cappuccino' AND i.name IN ('Coffee Beans', 'Whole Milk'))
    OR (m.name = 'Latte' AND i.name IN ('Coffee Beans', 'Whole Milk'));

-- Orders (at least 30 in different statuses)
DO $$
DECLARE
    customer_names TEXT[] := ARRAY['John Smith', 'Emma Davis', 'Michael Johnson', 'Sarah Wilson', 'David Brown', 'Lisa Anderson', 'James Taylor', 'Jennifer Martinez', 'Robert Garcia', 'Maria Rodriguez'];
    statuses order_status[] := ARRAY['pending', 'preparing', 'ready', 'delivered', 'cancelled']::order_status[];
    i INTEGER;
    selected_status order_status;
    selected_name TEXT;
    new_order_id UUID;
BEGIN
    FOR i IN 1..35 LOOP
        -- Select random status and name
        selected_status := statuses[1 + (i % 5)];
        selected_name := customer_names[1 + (i % 10)];
        
        -- Insert order
        INSERT INTO orders (customer_name, special_instructions, total_amount, status, created_at)
        VALUES (
            selected_name,
                                CASE WHEN i % 3 = 0 THEN '{"notes": "Extra hot"}'::JSONB ELSE NULL END,
            (5 + (i % 20))::DECIMAL(10,2),
            selected_status,
            CURRENT_TIMESTAMP - (i || ' days')::INTERVAL
        ) RETURNING order_id INTO new_order_id;

        -- Insert order status history
        INSERT INTO order_status_history (order_id, old_status, new_status, change_reason)
        VALUES (
            new_order_id,
            'pending',
            selected_status,
            CASE 
                WHEN selected_status = 'cancelled' THEN 'Customer request'
                WHEN selected_status = 'delivered' THEN 'Order completed'
                ELSE 'Regular processing'
            END
        );
    END LOOP;
END $$;

-- Order Items
INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
SELECT 
    o.order_id,
    m.menu_item_id,
    1 + (random() * 3)::INT,
    m.price,
    CASE 
        WHEN m.name = 'Latte' THEN '{"milk": "oat"}'::jsonb
        WHEN m.name = 'Cappuccino' THEN '{"extra_shot": true}'::jsonb
        ELSE NULL
    END
FROM orders o
CROSS JOIN menu_items m
WHERE o.order_id IN (SELECT order_id FROM orders LIMIT 35)
AND m.name IN ('Latte', 'Cappuccino', 'Espresso')
LIMIT 50;

-- Price History (spanning several months)
INSERT INTO price_history (menu_item_id, old_price, new_price, changed_at, change_reason)
SELECT 
    menu_item_id,
    price - 0.50,
    price,
    CURRENT_TIMESTAMP - (generate_series(1, 6) || ' months')::INTERVAL,
    'Regular price adjustment'
FROM menu_items
WHERE name IN ('Latte', 'Cappuccino', 'Espresso');

-- Inventory Transactions (showing stock movements)
INSERT INTO inventory_transactions (ingredient_id, quantity_change, transaction_type, reason, created_at)
SELECT 
    i.ingredient_id,
    CASE 
        WHEN it.type = 'addition' THEN 1000
        ELSE -500
    END,
    it.type::transaction_type,
    it.reason,
    CURRENT_TIMESTAMP - (it.days || ' days')::INTERVAL
FROM inventory i
CROSS JOIN (
    VALUES 
        ('addition', 'Weekly restock', 1),
        ('deduction', 'Daily usage', 2),
        ('adjustment', 'Inventory check', 3),
        ('waste', 'Expired products', 4)
) as it(type, reason, days)
WHERE i.name IN ('Coffee Beans', 'Whole Milk', 'Sugar')
LIMIT 50;

-- Additional indexes for specific query patterns
CREATE INDEX idx_inventory_transactions_date ON inventory_transactions(created_at);
CREATE INDEX idx_price_history_date ON price_history(changed_at);
CREATE INDEX idx_order_status_history_date ON order_status_history(changed_at);
CREATE INDEX idx_menu_items_name_price ON menu_items(name, price);
CREATE INDEX idx_inventory_name_quantity ON inventory(name, quantity);