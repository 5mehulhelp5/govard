-- STAGING environment initialization
CREATE TABLE IF NOT EXISTS core_config_data (
    config_id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(255) NOT NULL,
    value TEXT
);

INSERT INTO core_config_data (path, value) VALUES 
    ('general/store/name', 'STAGING Store'),
    ('test/environment', 'staging'),
    ('test/feature_flag', 'disabled');

CREATE TABLE IF NOT EXISTS test_markers (
    marker_id INT AUTO_INCREMENT PRIMARY KEY,
    env_name VARCHAR(50) NOT NULL,
    marker_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO test_markers (env_name, marker_value) VALUES 
    ('staging', CONCAT('STAGING_SEED_', UNIX_TIMESTAMP()));

-- Production-like product data
CREATE TABLE IF NOT EXISTS catalog_product_entity (
    entity_id INT AUTO_INCREMENT PRIMARY KEY,
    sku VARCHAR(64) NOT NULL,
    type_id VARCHAR(32) NOT NULL DEFAULT 'simple'
);

INSERT INTO catalog_product_entity (sku, type_id) VALUES 
    ('STAGING-PRODUCT-001', 'simple'),
    ('STAGING-PRODUCT-002', 'simple');
