-- schema.sql
CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    gulagged DATETIME DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL,
    command TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS sounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE CHECK (name <> ''),
    category_id INTEGER NOT NULL,
    hash TEXT NOT NULL UNIQUE,
    file BLOB NOT NULL,
    CONSTRAINT fk_categories
        FOREIGN KEY (category_id)
        REFERENCES categories(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS stats_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    sound_id INTEGER NOT NULL,
    count INTEGER DEFAULT 0,
    UNIQUE(user_id, sound_id),
    CONSTRAINT fk_sounds
        FOREIGN KEY (sound_id)
        REFERENCES sounds(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS user_favorites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    sound_id INTEGER NOT NULL,
    UNIQUE(user_id, sound_id),
    CONSTRAINT fk_sounds
        FOREIGN KEY (sound_id)
        REFERENCES sounds(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_users
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);