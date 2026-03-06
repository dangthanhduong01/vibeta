-- Khởi tạo database cho VibeTA chat application
CREATE DATABASE IF NOT EXISTS vibeta_chat;

-- Tạo extension cho UUID generation nếu cần
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Indexes để tối ưu performance
-- Sẽ được tạo tự động bởi GORM khi application khởi động
-- Nhưng có thể tạo thêm indexes custom ở đây nếu cần
