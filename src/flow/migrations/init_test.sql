CREATE TABLE IF NOT EXISTS users_test (
                                     id INTEGER PRIMARY KEY,
                                     nickname TEXT,
                                     paused BOOL
);

CREATE TABLE IF NOT EXISTS channels_test (
                                        nickname TEXT,
                                        channel TEXT,
                                        topic TEXT,
                                        last_time TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS messages_test (
                                        nickname TEXT,
                                        link TEXT,
                                        channel TEXT,
                                        topic TEXT,
                                        summary TEXT
);
