package model

const (
	// TODO SEEDMS replace this with own table definitions
	seedsTable = `
CREATE TABLE IF NOT EXISTS seed_data (
  id INT PRIMARY KEY NOT NULL,
  name STRING NOT NULL
);
`
)
