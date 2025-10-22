## rate_limit_rules table
CREATE TABLE rate_limiter_rules (
    id SERIAL PRIMARY KEY,
    action VARCHAR(50) NOT NULL,         -- tên hành động: post, like, comment, follow_unfollow, requests_per_ip...
    target_type VARCHAR(50) NOT NULL,    -- áp dụng cho: user, ip, global, post...
    limit_value INT NOT NULL,            -- số lượng tối đa
    time_unit VARCHAR(20) NOT NULL,      -- "second", "minute", "hour"
    description TEXT,                    -- mô tả rule
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

## check postgresql service status
# 1️⃣ Kiểm tra trạng thái PostgreSQL
sudo systemctl status postgresql
# 2️⃣ Khởi động PostgreSQL nếu cần
sudo systemctl start postgresql

## cmd to create db 
# 1️⃣ Kết nối vào PostgreSQL
sudo -u postgres psql

# 2️⃣ Tạo user
CREATE USER taopq WITH PASSWORD '123456a@';

# 3️⃣ Tạo database
CREATE DATABASE mydb OWNER taopq;

# Login to mydb if it created
psql -h localhost -U taopq -d mydb 

# 4️⃣ Cấp quyền cho user
GRANT ALL PRIVILEGES ON DATABASE mydb TO taopq;

## change owner db
ALTER TABLE public.rate_limiter_rules OWNER TO taopq;
ALTER TABLE public.users OWNER TO taopq;