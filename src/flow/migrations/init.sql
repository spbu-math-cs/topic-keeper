CREATE TABLE IF NOT EXISTS mm_chans (
    id TEXT PRIMARY KEY,
    name TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER, --PRIMARY KEY,
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


CREATE TABLE IF NOT EXISTS  vk_last_post_by_public (
    groupid TEXT PRIMARY KEY,
    last_post INT,
    public_name TEXT
)

