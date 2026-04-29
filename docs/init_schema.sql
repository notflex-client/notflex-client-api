-- ============================================================
--  STREAMING PLATFORM — FULL DATABASE SCHEMA
--  Database: PostgreSQL 14+
--  Project:  Subscription-based Movie Streaming + AI Recommendation
-- ============================================================

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- ============================================================
--  SECTION 1: USER & AUTHENTICATION
-- ============================================================

CREATE TYPE user_role AS ENUM ('admin', 'subscriber', 'guest');

CREATE TABLE users (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(100),
    avatar_url    TEXT,
    role          user_role   NOT NULL DEFAULT 'guest',
    is_active     BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP   NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  users               IS 'Tài khoản người dùng hệ thống';
COMMENT ON COLUMN users.role          IS 'admin = quản trị, subscriber = đã mua gói, guest = chưa mua';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hash, không lưu plain text';


-- ============================================================
--  SECTION 2: SUBSCRIPTION & PAYMENT
-- ============================================================

CREATE TABLE subscription_plans (
    id            SERIAL       PRIMARY KEY,
    name          VARCHAR(50)  NOT NULL,          -- 'Monthly', 'Annual'
    price         NUMERIC(10,2) NOT NULL,
    duration_days INT          NOT NULL,          -- 30 | 365
    description   TEXT,
    is_active     BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE subscription_plans IS 'Các gói đăng ký (Monthly / Annual)';


CREATE TYPE subscription_status AS ENUM ('active', 'expired', 'cancelled');

CREATE TABLE user_subscriptions (
    id         UUID                PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID                NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id    INT                 NOT NULL REFERENCES subscription_plans(id),
    start_date TIMESTAMP           NOT NULL DEFAULT NOW(),
    end_date   TIMESTAMP           NOT NULL,
    status     subscription_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMP           NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  user_subscriptions          IS 'Lịch sử đăng ký gói của từng user';
COMMENT ON COLUMN user_subscriptions.end_date IS 'Backend check cột này để validate quyền xem phim';


CREATE TYPE payment_status AS ENUM ('pending', 'success', 'failed', 'refunded');

CREATE TABLE payments (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID           NOT NULL REFERENCES users(id),
    subscription_id UUID           REFERENCES user_subscriptions(id),
    amount          NUMERIC(10,2)  NOT NULL,
    payment_method  VARCHAR(50),                 -- 'momo', 'vnpay', 'card'
    status          payment_status NOT NULL DEFAULT 'pending',
    transaction_id  VARCHAR(255),                -- mã giao dịch từ cổng thanh toán
    created_at      TIMESTAMP      NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  payments                IS 'Lịch sử giao dịch thanh toán';
COMMENT ON COLUMN payments.transaction_id IS 'Mã trả về từ MoMo / VNPay để đối soát';


-- ============================================================
--  SECTION 3: MOVIE & CONTENT
-- ============================================================

CREATE TABLE genres (
    id   SERIAL      PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL
);

COMMENT ON TABLE genres IS 'Thể loại phim: Action, Drama, Comedy, ...';


CREATE TABLE movies (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    title         VARCHAR(255)  NOT NULL,
    description   TEXT,
    poster_url    TEXT,
    video_url     TEXT,                          -- link HLS / CDN
    trailer_url   TEXT,
    duration_mins INT,
    release_year  SMALLINT,
    avg_rating    NUMERIC(3,1)  NOT NULL DEFAULT 0,  -- cập nhật khi có rating mới
    is_premium    BOOLEAN       NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMP     NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP     NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE  movies            IS 'Kho phim của nền tảng';
COMMENT ON COLUMN movies.video_url  IS 'URL stream thực tế (HLS .m3u8), chỉ trả về nếu user có subscription hợp lệ';
COMMENT ON COLUMN movies.is_premium IS 'TRUE = cần subscription, FALSE = xem miễn phí (trailer/demo)';


-- Quan hệ nhiều-nhiều: phim ↔ thể loại
CREATE TABLE movie_genres (
    movie_id UUID REFERENCES movies(id) ON DELETE CASCADE,
    genre_id INT  REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (movie_id, genre_id)
);


-- ============================================================
--  SECTION 4: AI RECOMMENDATION — DỮ LIỆU ĐẦU VÀO
-- ============================================================

-- Lịch sử xem — nguồn dữ liệu chính cho AI
CREATE TABLE watch_history (
    id             UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    movie_id       UUID      NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    watched_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    watch_duration INT       NOT NULL DEFAULT 0,   -- số giây đã xem
    is_completed   BOOLEAN   NOT NULL DEFAULT FALSE
);

COMMENT ON TABLE  watch_history                IS 'Mỗi lần user bấm play = 1 record; AI service đọc bảng này';
COMMENT ON COLUMN watch_history.watch_duration IS 'Giây đã xem — dùng để tính mức độ quan tâm (engagement weight)';
COMMENT ON COLUMN watch_history.is_completed   IS 'TRUE nếu user xem >= 90% thời lượng phim';


-- Đánh giá phim (1–5 sao)
CREATE TABLE movie_ratings (
    user_id  UUID     NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    movie_id UUID     NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    rating   SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    rated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, movie_id)
);

COMMENT ON TABLE movie_ratings IS '1 user chỉ rate 1 lần / 1 phim (PRIMARY KEY kép)';


-- ============================================================
--  SECTION 5: INDEXES — TỐI ƯU QUERY THƯỜNG GẶP
-- ============================================================

-- [Auth] Đăng nhập bằng email
CREATE INDEX idx_users_email ON users(email);

-- [Subscription] Check user có subscription active không — query nhiều nhất
CREATE INDEX idx_user_subscriptions_user_status
    ON user_subscriptions(user_id, status, end_date DESC);

-- [Payment] Xem lịch sử thanh toán theo user
CREATE INDEX idx_payments_user ON payments(user_id, created_at DESC);

-- [Movie] Tìm phim theo năm, lọc premium
CREATE INDEX idx_movies_release_year ON movies(release_year);
CREATE INDEX idx_movies_is_premium   ON movies(is_premium);

-- [Movie Genre] Lọc phim theo thể loại
CREATE INDEX idx_movie_genres_genre ON movie_genres(genre_id);

-- [Watch History] AI service + lịch sử xem theo user
CREATE INDEX idx_watch_history_user    ON watch_history(user_id, watched_at DESC);
CREATE INDEX idx_watch_history_movie   ON watch_history(movie_id);

-- [Rating] Tính avg_rating theo phim
CREATE INDEX idx_movie_ratings_movie ON movie_ratings(movie_id);


-- ============================================================
--  SECTION 6: TRIGGER — TỰ ĐỘNG CẬP cấu hình avg_rating
-- ============================================================

CREATE OR REPLACE FUNCTION update_movie_avg_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE movies
    SET avg_rating = (
        SELECT ROUND(AVG(rating)::NUMERIC, 1)
        FROM movie_ratings
        WHERE movie_id = COALESCE(NEW.movie_id, OLD.movie_id)
    )
    WHERE id = COALESCE(NEW.movie_id, OLD.movie_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_avg_rating
AFTER INSERT OR UPDATE OR DELETE ON movie_ratings
FOR EACH ROW EXECUTE FUNCTION update_movie_avg_rating();


-- ============================================================
--  SECTION 7: BẢNG BỔ SUNG (thêm sau — không có trigger/enum)
-- ============================================================

-- Token xác thực Bearer (1 user có thể có nhiều token — đa thiết bị)
CREATE TABLE user_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_agent TEXT,
    created_at TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_tokens_user ON user_tokens(user_id);

COMMENT ON TABLE user_tokens IS 'Bearer token lưu DB; xóa token = đăng xuất thiết bị đó';


-- Yêu cầu đăng ký chờ xác minh OTP (tự xóa sau khi dùng)
CREATE TABLE register_requests (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email             VARCHAR(255) NOT NULL,
    confirmation_code VARCHAR(6)  NOT NULL,
    verified          BOOLEAN     NOT NULL DEFAULT FALSE,
    expire_at         TIMESTAMP   NOT NULL,
    created_at        TIMESTAMP   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_register_requests_email ON register_requests(email);

COMMENT ON TABLE register_requests IS 'OTP đăng ký — expire_at = created_at + 5 phút; xóa sau khi confirm thành công';


-- ============================================================
--  SECTION 8: SEED DATA — DỮ LIỆU MẪU ĐỂ TEST
-- ============================================================

-- Gói subscription
INSERT INTO subscription_plans (name, price, duration_days, description) VALUES
    ('Monthly',  79000,  30,  'Gói tháng — xem không giới hạn trong 30 ngày'),
    ('Annual',  699000, 365,  'Gói năm — tiết kiệm 30% so với mua theo tháng');

-- Thể loại phim
INSERT INTO genres (name) VALUES
    ('Action'), ('Drama'), ('Comedy'), ('Thriller'),
    ('Sci-Fi'), ('Horror'), ('Romance'), ('Animation');

-- Tài khoản admin mẫu (password: Admin@123 — bcrypt hash)
INSERT INTO users (email, password_hash, full_name, role) VALUES
    ('admin@streamflix.com',
     '$2a$10$examplehashforAdminpassword00000000000000000',
     'System Admin', 'admin');
