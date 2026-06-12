-- +goose up
CREATE TABLE users(
  id INT PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  email TEXT UNIQUE NOT NULL
);
-- +goose Down
DROP TABLE users;
