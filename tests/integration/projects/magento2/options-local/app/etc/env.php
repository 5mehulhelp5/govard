<?php
return [
    'backend' => ['frontName' => 'admin'],
    'crypt' => ['key' => 'local-test-key-1234567890abcdef'],
    'db' => [
        'table_prefix' => '',
        'connection' => [
            'default' => [
                'host' => 'local-db',
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
    'MAGE_MODE' => 'developer',
    'session' => [
        'save' => 'files'
    ],
    'cache' => [
        'frontend' => [
            'default' => [
                'id_prefix' => 'local_'
            ],
            'page_cache' => [
                'id_prefix' => 'local_'
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
