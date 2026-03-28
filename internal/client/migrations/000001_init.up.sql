CREATE TABLE IF NOT EXISTS meta (
                                    id INTEGER PRIMARY KEY CHECK (id = 0),
    last_sync INTEGER NOT NULL DEFAULT 0,
    master_password_hash VARCHAR(250) NOT NULL DEFAULT '',
    token VARCHAR(250) NOT NULL DEFAULT ''
    );

INSERT OR IGNORE INTO meta (id, last_sync, master_password_hash, token)
VALUES (0, 0, '', '');

CREATE TABLE IF NOT EXISTS user_data (
                                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                                         user_id INTEGER NOT NULL DEFAULT 0,
                                         data_key TEXT NOT NULL,
                                         data_value BLOB NOT NULL,
                                         updated_at INTEGER NOT NULL,
                                         deleted_at INTEGER NOT NULL DEFAULT 0,
                                         UNIQUE(user_id, data_key)
    );