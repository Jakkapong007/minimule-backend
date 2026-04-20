-- Seed data for local development/testing
-- Run: docker exec -i minimule-backend-postgres-1 psql -U minimule -d minimule < migrations/seed.sql

-- All passwords are: password123
-- bcrypt hash for "password123" at cost 12
-- $2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RuK9gKFq6

-- ── Users ─────────────────────────────────────────────────────────────────────
INSERT INTO users (id, email, password_hash, phone, full_name, avatar_url, role, is_active) VALUES
  ('00000000-0000-0000-0000-000000000001', 'admin@minimule.com',    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RuK9gKFq6', '0800000001', 'Admin User',      NULL,                                     'admin',    TRUE),
  ('00000000-0000-0000-0000-000000000002', 'artist@minimule.com',   '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RuK9gKFq6', '0800000002', 'Nong Artist',     'https://picsum.photos/seed/artist/200', 'artist',   TRUE),
  ('00000000-0000-0000-0000-000000000003', 'customer@minimule.com', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/RuK9gKFq6', '0800000003', 'Somchai Customer','https://picsum.photos/seed/user/200',   'customer', TRUE);

-- ── User Profiles ─────────────────────────────────────────────────────────────
INSERT INTO user_profiles (id, user_id, bio, preferred_language) VALUES
  ('00000000-0000-0000-0001-000000000001', '00000000-0000-0000-0000-000000000001', 'Platform administrator.', 'en'),
  ('00000000-0000-0000-0001-000000000002', '00000000-0000-0000-0000-000000000002', 'Indie sticker artist based in Chiang Mai. I love cats and coffee.', 'th'),
  ('00000000-0000-0000-0001-000000000003', '00000000-0000-0000-0000-000000000003', 'Art lover and collector.', 'th');

-- ── User Addresses ────────────────────────────────────────────────────────────
INSERT INTO user_addresses (id, user_id, label, recipient_name, phone, address_line1, address_line2, province, postal_code, is_default) VALUES
  ('00000000-0000-0000-0002-000000000001', '00000000-0000-0000-0000-000000000003', 'Home',   'Somchai Customer', '0800000003', '123 Sukhumvit Soi 11', 'Khlong Toei Nuea', 'Bangkok',    '10110', TRUE),
  ('00000000-0000-0000-0002-000000000002', '00000000-0000-0000-0000-000000000003', 'Office', 'Somchai Customer', '0800000003', '456 Silom Road',       'Bang Rak',        'Bangkok',    '10500', FALSE),
  ('00000000-0000-0000-0002-000000000003', '00000000-0000-0000-0000-000000000002', 'Studio', 'Nong Artist',      '0800000002', '789 Nimman Road',      'Su Thep',         'Chiang Mai', '50200', TRUE);

-- ── Categories ────────────────────────────────────────────────────────────────
INSERT INTO categories (id, name, slug, icon_url, display_order, is_active) VALUES
  ('00000000-0000-0000-0003-000000000001', 'Stickers',      'stickers',       'https://picsum.photos/seed/sticker/64',    1, TRUE),
  ('00000000-0000-0000-0003-000000000002', 'Prints & Posters','prints-posters','https://picsum.photos/seed/print/64',     2, TRUE),
  ('00000000-0000-0000-0003-000000000003', 'Enamel Pins',   'enamel-pins',    'https://picsum.photos/seed/pin/64',        3, TRUE),
  ('00000000-0000-0000-0003-000000000004', 'Tote Bags',     'tote-bags',      'https://picsum.photos/seed/bag/64',        4, TRUE),
  ('00000000-0000-0000-0003-000000000005', 'Keychains',     'keychains',      'https://picsum.photos/seed/keychain/64',   5, TRUE);

-- ── Products ──────────────────────────────────────────────────────────────────
INSERT INTO products (id, name, description, base_price, stock_qty, is_featured, popularity_score, status, category_id, seller_id) VALUES
  ('00000000-0000-0000-0004-000000000001', 'Cat Ramen Sticker Pack',   'A pack of 6 die-cut vinyl stickers featuring cats eating ramen.',      129.00, 50, TRUE,  95.0, 'active', '00000000-0000-0000-0003-000000000001', '00000000-0000-0000-0000-000000000002'),
  ('00000000-0000-0000-0004-000000000002', 'Mushroom Forest Print',    'A3 art print of a whimsical mushroom forest, printed on 200gsm paper.',350.00, 20, TRUE,  80.0, 'active', '00000000-0000-0000-0003-000000000002', '00000000-0000-0000-0000-000000000002'),
  ('00000000-0000-0000-0004-000000000003', 'Space Corgi Enamel Pin',   'Hard enamel pin of a corgi in a space suit. 3cm, rubber clasp.',       189.00, 35, FALSE, 70.0, 'active', '00000000-0000-0000-0003-000000000003', '00000000-0000-0000-0000-000000000002'),
  ('00000000-0000-0000-0004-000000000004', 'Matcha Latte Tote Bag',    'Natural canvas tote with matcha latte illustration, screen printed.',   290.00, 15, TRUE,  88.0, 'active', '00000000-0000-0000-0003-000000000004', '00000000-0000-0000-0000-000000000002'),
  ('00000000-0000-0000-0004-000000000005', 'Shiba Inu Acrylic Keychain','Double-sided acrylic keychain featuring a Shiba Inu in pajamas.',      99.00,  60, FALSE, 60.0, 'active', '00000000-0000-0000-0003-000000000005', '00000000-0000-0000-0000-000000000002'),
  ('00000000-0000-0000-0004-000000000006', 'Draft Watercolor Set',     'Work in progress — not yet released.',                                  500.00,  0, FALSE,  0.0, 'draft',  '00000000-0000-0000-0003-000000000002', '00000000-0000-0000-0000-000000000002');

-- ── Product Images ────────────────────────────────────────────────────────────
INSERT INTO product_images (id, product_id, image_url, sort_order, is_primary) VALUES
  ('00000000-0000-0000-0005-000000000001', '00000000-0000-0000-0004-000000000001', 'https://picsum.photos/seed/p1a/600/600', 0, TRUE),
  ('00000000-0000-0000-0005-000000000002', '00000000-0000-0000-0004-000000000001', 'https://picsum.photos/seed/p1b/600/600', 1, FALSE),
  ('00000000-0000-0000-0005-000000000003', '00000000-0000-0000-0004-000000000002', 'https://picsum.photos/seed/p2a/600/600', 0, TRUE),
  ('00000000-0000-0000-0005-000000000004', '00000000-0000-0000-0004-000000000003', 'https://picsum.photos/seed/p3a/600/600', 0, TRUE),
  ('00000000-0000-0000-0005-000000000005', '00000000-0000-0000-0004-000000000004', 'https://picsum.photos/seed/p4a/600/600', 0, TRUE),
  ('00000000-0000-0000-0005-000000000006', '00000000-0000-0000-0004-000000000005', 'https://picsum.photos/seed/p5a/600/600', 0, TRUE);

-- ── Product Variants ──────────────────────────────────────────────────────────
INSERT INTO product_variants (id, product_id, sku, size, color, price_modifier, stock_qty) VALUES
  -- Sticker pack: no variants, skip
  -- Mushroom print: A3 / A4 sizes
  ('00000000-0000-0000-0006-000000000001', '00000000-0000-0000-0004-000000000002', 'MUSH-A3', 'A3', NULL, 0.00,   15),
  ('00000000-0000-0000-0006-000000000002', '00000000-0000-0000-0004-000000000002', 'MUSH-A4', 'A4', NULL, -80.00, 30),
  -- Tote bag: natural / black
  ('00000000-0000-0000-0006-000000000003', '00000000-0000-0000-0004-000000000004', 'TOTE-NAT', NULL, 'Natural', 0.00,  10),
  ('00000000-0000-0000-0006-000000000004', '00000000-0000-0000-0004-000000000004', 'TOTE-BLK', NULL, 'Black',   20.00,  5);

-- ── Carts ─────────────────────────────────────────────────────────────────────
INSERT INTO carts (id, user_id, status) VALUES
  ('00000000-0000-0000-0007-000000000001', '00000000-0000-0000-0000-000000000003', 'active');

-- ── Cart Items ────────────────────────────────────────────────────────────────
INSERT INTO cart_items (id, cart_id, product_id, variant_id, quantity, unit_price) VALUES
  ('00000000-0000-0000-0008-000000000001', '00000000-0000-0000-0007-000000000001', '00000000-0000-0000-0004-000000000001', NULL,                                     2, 129.00),
  ('00000000-0000-0000-0008-000000000002', '00000000-0000-0000-0007-000000000001', '00000000-0000-0000-0004-000000000003', NULL,                                     1, 189.00),
  ('00000000-0000-0000-0008-000000000003', '00000000-0000-0000-0007-000000000001', '00000000-0000-0000-0004-000000000004', '00000000-0000-0000-0006-000000000003',  1, 290.00);

-- ── Orders ────────────────────────────────────────────────────────────────────
INSERT INTO orders (id, order_number, user_id, address_id, subtotal, discount_amount, shipping_fee, total, status) VALUES
  ('00000000-0000-0000-0009-000000000001', 'MML-000001', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0002-000000000001', 538.00, 0.00, 50.00, 588.00, 'delivered'),
  ('00000000-0000-0000-0009-000000000002', 'MML-000002', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0002-000000000001', 350.00, 0.00, 50.00, 400.00, 'paid');

-- ── Order Items ───────────────────────────────────────────────────────────────
INSERT INTO order_items (id, order_id, product_id, variant_id, quantity, unit_price, subtotal) VALUES
  -- Order 1: sticker pack x2 + enamel pin x1
  ('00000000-0000-0000-0010-000000000001', '00000000-0000-0000-0009-000000000001', '00000000-0000-0000-0004-000000000001', NULL, 2, 129.00, 258.00),
  ('00000000-0000-0000-0010-000000000002', '00000000-0000-0000-0009-000000000001', '00000000-0000-0000-0004-000000000003', NULL, 1, 189.00, 189.00),
  -- Order 2: mushroom print A4
  ('00000000-0000-0000-0010-000000000003', '00000000-0000-0000-0009-000000000002', '00000000-0000-0000-0004-000000000002', '00000000-0000-0000-0006-000000000002', 1, 270.00, 270.00);

-- ── Posts ─────────────────────────────────────────────────────────────────────
INSERT INTO posts (id, user_id, caption, image_url, is_sticker_design, like_count, comment_count) VALUES
  ('00000000-0000-0000-0011-000000000001', '00000000-0000-0000-0000-000000000002', 'New sticker pack just dropped! 🐱🍜 Swipe to see all 6 designs.',          'https://picsum.photos/seed/post1/800/800', TRUE,  42, 5),
  ('00000000-0000-0000-0011-000000000002', '00000000-0000-0000-0000-000000000002', 'Mushroom forest print — A3 now available in the shop!',                    'https://picsum.photos/seed/post2/800/800', FALSE, 28, 2),
  ('00000000-0000-0000-0011-000000000003', '00000000-0000-0000-0000-000000000002', 'Working on something new... a whole underwater world 🌊',                  'https://picsum.photos/seed/post3/800/800', FALSE, 15, 1),
  ('00000000-0000-0000-0011-000000000004', '00000000-0000-0000-0000-000000000002', 'Space corgi enamel pin restocked! Only 35 left.',                          'https://picsum.photos/seed/post4/800/800', TRUE,   9, 0);

-- ── Post Likes ────────────────────────────────────────────────────────────────
INSERT INTO post_likes (post_id, user_id) VALUES
  ('00000000-0000-0000-0011-000000000001', '00000000-0000-0000-0000-000000000003'),
  ('00000000-0000-0000-0011-000000000002', '00000000-0000-0000-0000-000000000003'),
  ('00000000-0000-0000-0011-000000000003', '00000000-0000-0000-0000-000000000003');

-- ── Post Comments ─────────────────────────────────────────────────────────────
INSERT INTO post_comments (id, post_id, user_id, body) VALUES
  ('00000000-0000-0000-0012-000000000001', '00000000-0000-0000-0011-000000000001', '00000000-0000-0000-0000-000000000003', 'These are so cute!! Just ordered 2 packs 🥰'),
  ('00000000-0000-0000-0012-000000000002', '00000000-0000-0000-0011-000000000001', '00000000-0000-0000-0000-000000000001', 'Amazing work as always!'),
  ('00000000-0000-0000-0012-000000000003', '00000000-0000-0000-0011-000000000002', '00000000-0000-0000-0000-000000000003', 'The colors are gorgeous 😍');

-- ── Shipping Methods ──────────────────────────────────────────────────────────
INSERT INTO shipping_methods (id, name, description, carrier, estimated_days_min, estimated_days_max, base_fee, is_active) VALUES
  ('00000000-0000-0000-0013-000000000001', 'Standard Shipping',  'Regular delivery by Thailand Post',        'Thailand Post', 3, 7, 50.00,  TRUE),
  ('00000000-0000-0000-0013-000000000002', 'Express Shipping',   'Next-day delivery via Flash Express',      'Flash Express', 1, 2, 120.00, TRUE),
  ('00000000-0000-0000-0013-000000000003', 'Economy Shipping',   'Budget option, delivered in 5–10 days',   'Kerry Express', 5,10, 0.00,   TRUE);

-- ── Promotions ────────────────────────────────────────────────────────────────
INSERT INTO promotions (id, code, description, discount_type, discount_value, min_order_amount, max_uses, starts_at, expires_at, is_active) VALUES
  ('00000000-0000-0000-0014-000000000001', 'WELCOME10', 'Welcome discount — 10% off your first order', 'percentage', 10.00, 0.00,   1000, NOW() - INTERVAL '7 days', NOW() + INTERVAL '30 days', TRUE),
  ('00000000-0000-0000-0014-000000000002', 'FLAT50',    'Flat ฿50 off orders over ฿300',               'fixed',      50.00, 300.00, 500,  NOW() - INTERVAL '1 day',  NOW() + INTERVAL '14 days', TRUE),
  ('00000000-0000-0000-0014-000000000003', 'ARTIST20',  '20% off all prints (artist exclusive)',       'percentage', 20.00, 0.00,    100,  NOW() - INTERVAL '3 days', NOW() + INTERVAL '60 days', TRUE);

-- ── Product Reviews ───────────────────────────────────────────────────────────
INSERT INTO product_reviews (id, product_id, user_id, order_id, rating, body) VALUES
  ('00000000-0000-0000-0015-000000000001', '00000000-0000-0000-0004-000000000001', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0009-000000000001', 5, 'Absolutely love these stickers! Great quality vinyl, colors are vivid. Highly recommend.'),
  ('00000000-0000-0000-0015-000000000002', '00000000-0000-0000-0004-000000000003', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0009-000000000001', 4, 'Really cute pin. The enamel is smooth with no bubbles. Shipping was fast too!'),
  ('00000000-0000-0000-0015-000000000003', '00000000-0000-0000-0004-000000000002', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0009-000000000002', 5, 'Print quality is exceptional. Hung it in my studio immediately. Worth every baht.');

-- ── Notifications ─────────────────────────────────────────────────────────────
INSERT INTO notifications (id, user_id, type, title, body, is_read) VALUES
  ('00000000-0000-0000-0016-000000000001', '00000000-0000-0000-0000-000000000003', 'order_update', 'Order Delivered!',        'Your order MML-000001 has been delivered. Enjoy your goodies!', TRUE),
  ('00000000-0000-0000-0016-000000000002', '00000000-0000-0000-0000-000000000003', 'order_update', 'Order Confirmed',         'Your order MML-000002 has been paid and is being processed.',  FALSE),
  ('00000000-0000-0000-0016-000000000003', '00000000-0000-0000-0000-000000000003', 'promotion',    'New Promo: FLAT50',       'Use code FLAT50 to get 50 off orders over 300. Limited time!', FALSE),
  ('00000000-0000-0000-0016-000000000004', '00000000-0000-0000-0000-000000000002', 'post_vote',    'Someone loved your post!','Your post got 10 new votes. Keep creating!',                   FALSE),
  ('00000000-0000-0000-0016-000000000005', '00000000-0000-0000-0000-000000000003', 'system',       'Welcome to miniMule!',    'Discover stickers, prints, and unique artwork from indie artists.', TRUE);

-- ── Post Votes ────────────────────────────────────────────────────────────────
INSERT INTO post_votes (post_id, user_id) VALUES
  ('00000000-0000-0000-0011-000000000001', '00000000-0000-0000-0000-000000000003'),
  ('00000000-0000-0000-0011-000000000002', '00000000-0000-0000-0000-000000000003');

-- Update vote_count denorm
UPDATE posts SET vote_count = 1 WHERE id IN ('00000000-0000-0000-0011-000000000001','00000000-0000-0000-0011-000000000002');

-- ── User Payment Methods ──────────────────────────────────────────────────────
INSERT INTO user_payment_methods (id, user_id, type, label, brand, last_four, is_default) VALUES
  ('00000000-0000-0000-0017-000000000001', '00000000-0000-0000-0000-000000000003', 'credit_card', 'Visa •••• 4242', 'Visa', '4242', TRUE),
  ('00000000-0000-0000-0017-000000000002', '00000000-0000-0000-0000-000000000003', 'promptpay',   'PromptPay',      NULL,   NULL,   FALSE);

-- ── Search History ────────────────────────────────────────────────────────────
INSERT INTO search_history (id, user_id, query) VALUES
  ('00000000-0000-0000-0018-000000000001', '00000000-0000-0000-0000-000000000003', 'cat sticker'),
  ('00000000-0000-0000-0018-000000000002', '00000000-0000-0000-0000-000000000003', 'mushroom print'),
  ('00000000-0000-0000-0018-000000000003', '00000000-0000-0000-0000-000000000003', 'enamel pin');
