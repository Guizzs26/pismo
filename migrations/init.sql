CREATE TABLE IF NOT EXISTS accounts (
  account_id SERIAL PRIMARY KEY,
  document_number VARCHAR(20) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS operation_types (
  operation_type_id INT PRIMARY KEY,
  description VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
  transaction_id SERIAL PRIMARY KEY,
  account_id INT NOT NULL,
  operation_type_id INT NOT NULL,
  amount NUMERIC(15, 2) NOT NULL,
  event_date TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
  idempotency_key VARCHAR(64) UNIQUE,
  CONSTRAINT fk_account FOREIGN KEY(account_id) REFERENCES accounts(account_id),
  CONSTRAINT fk_operation_type FOREIGN KEY(operation_type_id) REFERENCES operation_types(operation_type_id)
);

INSERT INTO operation_types (operation_type_id, description) VALUES
  (1, 'PURCHASE'),
  (2, 'INSTALLMENT PURCHASE'),
  (3, 'WITHDRAWAL'),
  (4, 'PAYMENT')
ON CONFLICT (operation_type_id) DO NOTHING;