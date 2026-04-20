-- minimule-backend initial schema
-- Run: make migrate-up

-- ── Extensions ────────────────────────────────────────────────────────────────
CREATE EXTENSION IF NOT EXISTS "pgcrypto";   -- gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pg_trgm";    -- ILIKE search index

-- ── Sequences ─────────────────────────────────────────────────────────────────
CREATE SEQUENCE IF NOT EXISTS order_number_seq START 1;

-- ── Users ─────────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    phone         TEXT,
    full_name     TEXT,
    avatar_url    TEXT,
    role          TEXT        NOT NULL DEFAULT 'customer'
                              CHECK (role IN ('customer', 'artist', 'admin')),
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ
);

CREATE INDEX idx_users_email    ON users (email);
CREATE INDEX idx_users_role     ON users (role);

CREATE TABLE user_profiles (
    id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    bio                TEXT,
    date_of_birth      DATE,
    gender             TEXT,
    preferred_language TEXT        NOT NULL DEFAULT 'th',
    notification_push  BOOLEAN     NOT NULL DEFAULT TRUE,
    notification_email BOOLEAN     NOT NULL DEFAULT TRUE,
    notification_sms   BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE TABLE user_addresses (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label          TEXT,
    recipient_name TEXT        NOT NULL,
    phone          TEXT        NOT NULL,
    address_line1  TEXT        NOT NULL,
    address_line2  TEXT,
    subdistrict    TEXT,
    district       TEXT,
    province       TEXT,
    postal_code    TEXT,
    is_default     BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_user_addresses_user ON user_addresses (user_id);

-- ── Categories ────────────────────────────────────────────────────────────────
CREATE TABLE categories (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id     UUID        REFERENCES categories(id),
    name          TEXT        NOT NULL,
    slug          TEXT        UNIQUE,
    icon_url      TEXT,
    display_order INT         NOT NULL DEFAULT 0,
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON categories (parent_id);

-- ── Products ──────────────────────────────────────────────────────────────────
CREATE TABLE products (
    id               UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT          NOT NULL,
    description      TEXT,
    base_price       NUMERIC(12,2) NOT NULL CHECK (base_price >= 0),
    stock_qty        INT           NOT NULL DEFAULT 0 CHECK (stock_qty >= 0),
    is_customizable  BOOLEAN       NOT NULL DEFAULT FALSE,
    is_featured      BOOLEAN       NOT NULL DEFAULT FALSE,
    popularity_score FLOAT         NOT NULL DEFAULT 0,
    status           TEXT          NOT NULL DEFAULT 'draft'
                                   CHECK (status IN ('draft', 'active', 'archived')),
    category_id      UUID          NOT NULL REFERENCES categories(id),
    seller_id        UUID          NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ
);

CREATE INDEX idx_products_category   ON products (category_id);
CREATE INDEX idx_products_seller     ON products (seller_id);
CREATE INDEX idx_products_status     ON products (status);
CREATE INDEX idx_products_featured   ON products (is_featured) WHERE is_featured = TRUE;
CREATE INDEX idx_products_name_trgm  ON products USING GIN (name gin_trgm_ops);

CREATE TABLE product_images (
    id         UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID    NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url  TEXT    NOT NULL,
    alt_text   TEXT,
    sort_order INT     NOT NULL DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_product_images_product ON product_images (product_id);

CREATE TABLE product_variants (
    id             UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id     UUID          NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    sku            TEXT,
    size           TEXT,
    color          TEXT,
    material       TEXT,
    price_modifier NUMERIC(12,2) NOT NULL DEFAULT 0,
    stock_qty      INT           NOT NULL DEFAULT 0 CHECK (stock_qty >= 0)
);

CREATE INDEX idx_product_variants_product ON product_variants (product_id);

-- ── Carts ─────────────────────────────────────────────────────────────────────
CREATE TABLE carts (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        REFERENCES users(id) ON DELETE CASCADE,
    session_token TEXT,
    status        TEXT        NOT NULL DEFAULT 'active'
                              CHECK (status IN ('active', 'checked_out', 'abandoned')),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_carts_user_active ON carts (user_id) WHERE status = 'active';

CREATE TABLE cart_items (
    id        UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id   UUID          NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
    product_id UUID         NOT NULL REFERENCES products(id),
    variant_id UUID         REFERENCES product_variants(id),
    quantity  INT           NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0)
);

CREATE UNIQUE INDEX idx_cart_items_unique ON cart_items (cart_id, product_id, COALESCE(variant_id::TEXT, ''));

CREATE INDEX idx_cart_items_cart ON cart_items (cart_id);

-- ── Orders ────────────────────────────────────────────────────────────────────
CREATE TABLE orders (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number    TEXT          NOT NULL UNIQUE,
    user_id         UUID          NOT NULL REFERENCES users(id),
    address_id      UUID          NOT NULL REFERENCES user_addresses(id),
    subtotal        NUMERIC(12,2) NOT NULL,
    discount_amount NUMERIC(12,2) NOT NULL DEFAULT 0,
    shipping_fee    NUMERIC(12,2) NOT NULL DEFAULT 0,
    total           NUMERIC(12,2) NOT NULL,
    status          TEXT          NOT NULL DEFAULT 'pending'
                                  CHECK (status IN ('pending','paid','processing','shipped','delivered','cancelled','refunded')),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ
);

CREATE INDEX idx_orders_user   ON orders (user_id);
CREATE INDEX idx_orders_status ON orders (status);

CREATE TABLE order_items (
    id         UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID          NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id UUID          NOT NULL REFERENCES products(id),
    variant_id UUID          REFERENCES product_variants(id),
    quantity   INT           NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL,
    subtotal   NUMERIC(12,2) NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items (order_id);

-- ── updated_at trigger ────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at      BEFORE UPDATE ON users      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_products_updated_at   BEFORE UPDATE ON products   FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_carts_updated_at      BEFORE UPDATE ON carts      FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_orders_updated_at     BEFORE UPDATE ON orders     FOR EACH ROW EXECUTE FUNCTION set_updated_at();
