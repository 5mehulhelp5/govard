-- DEV environment initialization
CREATE TABLE IF NOT EXISTS core_config_data (
    config_id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(255) NOT NULL,
    value TEXT
);

INSERT INTO core_config_data (path, value) VALUES 
    ('general/store/name', 'DEV Store'),
    ('test/environment', 'dev'),
    ('test/feature_flag', 'enabled');

CREATE TABLE IF NOT EXISTS test_markers (
    marker_id INT AUTO_INCREMENT PRIMARY KEY,
    env_name VARCHAR(50) NOT NULL,
    marker_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_markers (env_name, marker_value) VALUES 
    ('dev', CONCAT('DEV_SEED_', UNIX_TIMESTAMP())),
    ('dev', 'DEV_DATA_001'),
    ('dev', 'DEV_DATA_002');

-- Sample product data
CREATE TABLE IF NOT EXISTS catalog_product_entity (
    entity_id INT AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(64) NOT NULL,
    type_id VARCHAR(32) NOT NULL DEFAULT 'simple'
);

INSERT INTO catalog_product_entity (sku, type_id) VALUES 
    ('DEV-PRODUCT-001', 'simple'),
    ('DEV-PRODUCT-002', 'configurable'),
    ('DEV-PRODUCT-003', 'simple'),
    ('DEV-PRODUCT-004', 'bundle');

-- Sample customers
CREATE TABLE IF NOT EXISTS customer_entity (
    entity_id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    firstname VARCHAR(255),
    lastname VARCHAR(255)
);

INSERT INTO customer_entity (email, firstname, lastname) VALUES 
    ('dev1@example.com', 'Dev', 'User1'),
    ('dev2@example.com', 'Dev', 'User2');
