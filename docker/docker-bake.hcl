variable "DOCKER_ORG" {
  default = "ddtcorex/govard-"
}

group "default" {
  targets = [
    "apache",
    "elasticsearch",
    "mariadb",
    "mysql",
    "nginx",
    "opensearch",
    "php",
    "php-magento2",
    "php-debug",
    "php-magento2-debug",
    "rabbitmq",
    "redis",
    "valkey",
    "varnish",
    "dnsmasq",
  ]
}

# ─── Apache ────────────────────────────────────────────────────────────────
target "apache" {
  name    = "apache-${replace(version, ".", "-")}"
  context = "docker/apache"
  matrix = {
    version = ["2.4", "latest"]
  }
  args = {
    APACHE_VERSION = "2.4.66"
  }
  tags = ["${DOCKER_ORG}apache:${version}"]
}

# ─── Elasticsearch ─────────────────────────────────────────────────────────
target "elasticsearch" {
  name       = "elasticsearch-${replace(version, ".", "-")}"
  context    = "docker/elasticsearch"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["8.15", "7.17", "7.16", "7.10", "7.9", "7.7", "7.6", "6.8", "5.6", "2.4"]
  }
  args = {
    ELASTICSEARCH_VERSION = (
      version == "8.15" ? "8.15.0" :
      version == "7.17" ? "7.17.10" :
      version == "7.16" ? "7.16.3" :
      version == "7.10" ? "7.10.2" :
      version == "7.9"  ? "7.9.3" :
      version == "7.7"  ? "7.7.1" :
      version == "7.6"  ? "7.6.2" :
      version == "6.8"  ? "6.8.23" :
      version == "5.6"  ? "5.6.16" :
      version == "2.4"  ? "2.4.6" : version
    )
    ELASTICSEARCH_IMAGE = version == "2.4" ? "elasticsearch" : "docker.elastic.co/elasticsearch/elasticsearch"
  }
  tags = ["${DOCKER_ORG}elasticsearch:${version}"]
}

# ─── MariaDB ───────────────────────────────────────────────────────────────
target "mariadb" {
  name       = "mariadb-${replace(version, ".", "-")}"
  context    = "docker/mariadb"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["11.4", "10.11", "10.6", "10.5", "10.4", "10.3", "10.2", "10.1", "10.0"]
  }
  args = {
    MARIADB_VERSION = version
  }
  tags = ["${DOCKER_ORG}mariadb:${version}"]
}

# ─── MySQL ─────────────────────────────────────────────────────────────────
target "mysql" {
  name       = "mysql-${replace(version, ".", "-")}"
  context    = "docker/mysql"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["8.4", "8.0", "5.7"]
  }
  args = {
    MYSQL_VERSION = version
  }
  tags = ["${DOCKER_ORG}mysql:${version}"]
}

# ─── Nginx ─────────────────────────────────────────────────────────────────
target "nginx" {
  name    = "nginx-${replace(version, ".", "-")}"
  context = "docker/nginx"
  matrix = {
    version = ["1.28", "latest"]
  }
  args = {
    NGINX_VERSION = "1.28.0"
  }
  tags = ["${DOCKER_ORG}nginx:${version}"]
}

# ─── OpenSearch ────────────────────────────────────────────────────────────
target "opensearch" {
  name       = "opensearch-${replace(version, ".", "-")}"
  context    = "docker/opensearch"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["3.0", "2.19", "2.12", "2.5", "1.3", "1.2"]
  }
  args = {
    OPENSEARCH_VERSION = (
      version == "3.0"  ? "3.0.0" :
      version == "2.19" ? "2.19.0" :
      version == "2.12" ? "2.12.0" :
      version == "2.5"  ? "2.5.0" :
      version == "1.3"  ? "1.3.20" :
      version == "1.2"  ? "1.2.0" : version
    )
  }
  tags = ["${DOCKER_ORG}opensearch:${version}"]
}

# ─── PHP ───────────────────────────────────────────────────────────────────
target "php" {
  name       = "php-${replace(version, ".", "-")}"
  context    = "docker/php"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["8.5", "8.4", "8.3", "8.2", "8.1", "7.4", "7.3", "7.2", "7.1", "7.0", "5.6"]
  }
  args = {
    PHP_VERSION = version
  }
  tags = ["${DOCKER_ORG}php:${version}"]
}

# ─── PHP (Magento 2) ───────────────────────────────────────────────────────
target "php-magento2" {
  name       = "php-magento2-${replace(version, ".", "-")}"
  context    = "docker/php"
  dockerfile = "magento2/Dockerfile"
  matrix = {
    version = ["8.5", "8.4", "8.3", "8.2", "8.1", "7.4", "7.3", "7.2"]
  }
  contexts = {
    "${DOCKER_ORG}php:${version}" = "target:php-${replace(version, ".", "-")}"
  }
  args = {
    PHP_VERSION             = version
    GOVARD_IMAGE_REPOSITORY = DOCKER_ORG
  }
  tags = ["${DOCKER_ORG}php-magento2:${version}"]
}

# ─── PHP (Debug) ───────────────────────────────────────────────────────────
target "php-debug" {
  name       = "php-debug-${replace(version, ".", "-")}"
  context    = "docker/php"
  dockerfile = "debug/Dockerfile"
  matrix = {
    version = ["8.5", "8.4", "8.3", "8.2", "8.1", "7.4", "7.3", "7.2", "7.1", "7.0", "5.6"]
  }
  args = {
    BASE_IMAGE = "${DOCKER_ORG}php:${version}"
  }
  contexts = {
    "${DOCKER_ORG}php:${version}" = "target:php-${replace(version, ".", "-")}"
  }
  tags = ["${DOCKER_ORG}php:${version}-debug"]
}

# ─── PHP (Magento 2 - Debug) ───────────────────────────────────────────────
target "php-magento2-debug" {
  name       = "php-magento2-debug-${replace(version, ".", "-")}"
  context    = "docker/php"
  dockerfile = "debug/Dockerfile"
  matrix = {
    version = ["8.5", "8.4", "8.3", "8.2", "8.1", "7.4", "7.3", "7.2"]
  }
  args = {
    BASE_IMAGE = "${DOCKER_ORG}php-magento2:${version}"
  }
  contexts = {
    "${DOCKER_ORG}php-magento2:${version}" = "target:php-magento2-${replace(version, ".", "-")}"
  }
  tags = ["${DOCKER_ORG}php-magento2:${version}-debug"]
}

# ─── RabbitMQ ──────────────────────────────────────────────────────────────
target "rabbitmq" {
  name       = "rabbitmq-${replace(version, ".", "-")}"
  context    = "docker/rabbitmq"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["4.1", "3.13", "3.12", "3.11", "3.9", "3.8", "3.7"]
  }
  args = {
    RABBITMQ_VERSION = version
  }
  tags = ["${DOCKER_ORG}rabbitmq:${version}"]
}

# ─── Redis ─────────────────────────────────────────────────────────────────
target "redis" {
  name       = "redis-${replace(version, ".", "-")}"
  context    = "docker/redis"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["7.4", "7.2", "7.0", "6.2", "6.0", "5.0", "4.0", "3.2"]
  }
  args = {
    REDIS_VERSION = version
  }
  tags = ["${DOCKER_ORG}redis:${version}"]
}

# ─── Valkey ────────────────────────────────────────────────────────────────
target "valkey" {
  name       = "valkey-${replace(version, ".", "-")}"
  context    = "docker/valkey"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["8.0", "7.2"]
  }
  args = {
    VALKEY_VERSION = version
  }
  tags = ["${DOCKER_ORG}valkey:${version}"]
}

# ─── Varnish ───────────────────────────────────────────────────────────────
target "varnish" {
  name       = "varnish-${replace(version, ".", "-")}"
  context    = "docker/varnish"
  dockerfile = "Dockerfile"
  matrix = {
    version = ["7.6", "7.4", "7.0", "6.0", "latest"]
  }
  args = {
    VARNISH_VERSION   = version == "latest" ? "7.6" : version
    VARNISH_IMAGE_TAG = version == "6.0" ? "6.0" : version == "latest" ? "7.6" : version
  }
  tags = ["${DOCKER_ORG}varnish:${version}"]
}

# ─── Dnsmasq ───────────────────────────────────────────────────────────────
target "dnsmasq" {
  context = "docker/dnsmasq"
  tags    = ["${DOCKER_ORG}dnsmasq:latest"]
}
