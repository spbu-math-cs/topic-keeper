CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    nickname TEXT,
    paused BOOL
);

CREATE TABLE IF NOT EXISTS channels (
    nickname TEXT,
    channel TEXT,
    topic TEXT,
    last_time TIMESTAMP WITH TIME ZONE,
    application TEXT
);

CREATE TABLE IF NOT EXISTS messages (
    nickname TEXT,
    link TEXT,
    channel TEXT,
    topic TEXT,
    summary TEXT,
    application TEXT
);
