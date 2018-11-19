-- +migrate Up
CREATE TABLE app_user (
  id VARCHAR(50) PRIMARY KEY,
  email VARCHAR(100),
  password VARCHAR(255),
  salt VARCHAR(255),
  locked BOOLEAN DEFAULT FALSE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(email)
);

CREATE TABLE session (
  id VARCHAR(50) PRIMARY KEY,
  user_id VARCHAR(50) REFERENCES app_user(id),
  created_at DATETIME
);

CREATE TABLE one_time_credential (
  id VARCHAR(50) PRIMARY KEY,
  key VARCHAR(255),
  user_id VARCHAR(50) REFERENCES app_user(id),
  has_been_used BOOLEAN DEFAULT FALSE,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  valid_to DATETIME
);

CREATE TABLE stock (
  symbol VARCHAR(50) PRIMARY KEY,
  name VARCHAR(100),
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE watchlist (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100),
    user_id VARCHAR(50) REFERENCES app_user(id),
    UNIQUE(name, user_id)
);

CREATE TABLE watchlist_member (
  symbol VARCHAR(50) REFERENCES stock(symbol),
  watchlist_id VARCHAR(50) REFERENCES watchlist(id),
  created_at DATETIME,
  PRIMARY KEY (symbol, watchlist_id)
);

-- +migrate Down
DROP TABLE IF EXISTS one_time_credential;
DROP TABLE IF EXISTS session;
DROP TABLE IF EXISTS watchlist_member;
DROP TABLE IF EXISTS watchlist;
DROP TABLE IF EXISTS stock;
DROP TABLE IF EXISTS user;