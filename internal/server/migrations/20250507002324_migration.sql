-- +goose Up
CREATE TABLE user (
    id INT AUTO_INCREMENT PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    master_password_hash VARCHAR(255) NOT NULL,
    UNIQUE KEY (login)
);

CREATE TABLE user_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    data_key VARCHAR(255) NOT NULL,
    data_value LONGBLOB NOT NULL,
    version INT NOT NULL DEFAULT 1,
    srv_updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES user(id),
    UNIQUE KEY (user_id, data_key)
);

CREATE INDEX idx_updated_at ON user_data(updated_at);
CREATE INDEX idx_deleted_at ON user_data(deleted_at);

-- +goose Down
DROP INDEX idx_deleted_at;
DROP INDEX idx_updated_at;
DROP TABLE user_data;
DROP TABLE user;
