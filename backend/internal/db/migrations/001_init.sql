CREATE TABLE users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  role TEXT NOT NULL DEFAULT 'affiliate' CHECK (role IN ('admin', 'seller', 'affiliate')),
  referral_code TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
  id BIGSERIAL PRIMARY KEY,
  seller_id BIGINT NOT NULL REFERENCES users(id),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
  commission_percent NUMERIC(5, 2) NOT NULL CHECK (commission_percent >= 0 AND commission_percent <= 100),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE clicks (
  id BIGSERIAL PRIMARY KEY,
  referral_code TEXT NOT NULL,
  product_id BIGINT NOT NULL REFERENCES products(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE conversions (
  id BIGSERIAL PRIMARY KEY,
  affiliate_id BIGINT REFERENCES users(id),
  seller_id BIGINT NOT NULL REFERENCES users(id),
  product_id BIGINT NOT NULL REFERENCES products(id),
  amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
  commission_amount_cents BIGINT NOT NULL CHECK (commission_amount_cents >= 0),
  payment_reference TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_seller_id ON products(seller_id);
CREATE INDEX idx_clicks_referral_code ON clicks(referral_code);
CREATE INDEX idx_clicks_product_id ON clicks(product_id);
CREATE INDEX idx_conversions_affiliate_id ON conversions(affiliate_id);
CREATE INDEX idx_conversions_seller_id ON conversions(seller_id);
CREATE INDEX idx_conversions_product_id ON conversions(product_id);
