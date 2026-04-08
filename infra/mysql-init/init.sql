SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- user table
CREATE TABLE IF NOT EXISTS `user` (
    `user_id` varchar(255) NOT NULL PRIMARY KEY,
    `username` varchar(255) NOT NULL,
    `email` varchar(255) NOT NULL,
    `password` varchar(255) NOT NULL,
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 