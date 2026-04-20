ALTER TABLE orders DROP COLUMN IF EXISTS promotion_code;
ALTER TABLE orders DROP COLUMN IF EXISTS promotion_id;
ALTER TABLE orders DROP COLUMN IF EXISTS shipping_method_id;

DROP TRIGGER IF EXISTS trg_post_vote_count ON post_votes;
DROP FUNCTION IF EXISTS update_post_vote_count();
ALTER TABLE posts DROP COLUMN IF EXISTS category_id;
ALTER TABLE posts DROP COLUMN IF EXISTS visibility;
ALTER TABLE posts DROP COLUMN IF EXISTS vote_count;
DROP TABLE IF EXISTS post_votes;

DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS promotions;

DROP TRIGGER IF EXISTS trg_product_rating ON product_reviews;
DROP FUNCTION IF EXISTS update_product_rating();
ALTER TABLE products DROP COLUMN IF EXISTS review_count;
ALTER TABLE products DROP COLUMN IF EXISTS avg_rating;
DROP TABLE IF EXISTS product_reviews;

DROP TABLE IF EXISTS shipments;
DROP TABLE IF EXISTS shipping_methods;
DROP TABLE IF EXISTS user_payment_methods;
