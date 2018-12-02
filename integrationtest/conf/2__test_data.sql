-- +migrate Up
INSERT INTO stock(symbol, name, created_at) VALUES
    ('AAPL', 'Apple inc.', CURRENT_TIMESTAMP),
    ('GOOG', 'Alphabet inc.', CURRENT_TIMESTAMP),
    ('NFLX', 'Netflix', CURRENT_TIMESTAMP),
    ('TSLA', 'Telsa', CURRENT_TIMESTAMP),
    ('FB', 'Facebook inc.', CURRENT_TIMESTAMP),
    ('AMZN', 'Amazon.com', CURRENT_TIMESTAMP);

-- +migrate Down
DELETE FROM stock 
WHERE symbol IN (
    'AAPL',
    'GOOG',
    'NFLX',
    'TSLA',
    'FB',
    'AMZN'
);