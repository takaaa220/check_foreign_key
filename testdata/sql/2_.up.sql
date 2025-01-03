	CREATE TABLE abs (
		id INT PRIMARY KEY,
		name VARCHAR(40) NOT NULL,
		age SMALLINT UNSIGNED DEFAULT 0,
		address_id BIGINT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		INDEX name_index (name),
		FOREIGN KEY (address_id) REFERENCES users (id)
	);
