CREATE TABLE IF NOT EXISTS `user` (
                                      id INT AUTO_INCREMENT PRIMARY KEY,
                                      login VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    master_password_hash VARCHAR(255) NOT NULL,
    UNIQUE KEY uk_user_login (login)
    );

CREATE TABLE IF NOT EXISTS user_data (
                                         id INT AUTO_INCREMENT PRIMARY KEY,
                                         user_id INT NOT NULL,
                                         data_key VARCHAR(255) NOT NULL,
    data_value LONGBLOB NOT NULL,
    srv_updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL,
    deleted_at DATETIME NULL,
    CONSTRAINT fk_user_data_user FOREIGN KEY (user_id) REFERENCES `user`(id),
    UNIQUE KEY uk_user_data_user_key (user_id, data_key)
    );

CREATE INDEX idx_user_data_updated_at ON user_data(updated_at);
CREATE INDEX idx_user_data_deleted_at ON user_data(deleted_at);