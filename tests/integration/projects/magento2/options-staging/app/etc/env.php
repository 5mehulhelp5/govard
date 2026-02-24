<?php
return [
    'backend' => ['frontName' => 'admin'],
    'crypt' => ['key' => 'staging-test-key-1234567890abcdef'],
    'db' => [
        'table_prefix' => '',
        'connection' => [
            'default' => [
                'host' => 'staging-db',
                'dbname' => 'magento',
                'username' => 'magento',
                'password' => 'magento',
                'active' => '1'
            ]
        ]
    ],
    'resource' => [
        'default_setup' => [
            'connection' => 'default'
        ]
    ],
    'x-frame-options' => 'SAMEORIGIN',
    'MAGE_MODE' => 'production',
    'session' => [
        'save' => 'files'
    ],
    'cache' => [
        'frontend' => [
            'default' => [
                'id_prefix' => 'staging_'
            ],
            'page_cache' => [
                'id_prefix' => 'staging_'
            ]
        ]
    ],
    'lock' => [
        'provider' => 'database'
    ],
    'directories' => [
        'document_root_is_pub' => true
    ]
];
