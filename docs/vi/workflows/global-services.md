---
title: Dịch vụ toàn cục
---

# Dịch vụ toàn cục (Global Services)

Govard cung cấp bộ dịch vụ tích hợp sẵn, được chia sẻ giữa tất cả các dự án. Các dịch vụ này chạy dưới dạng Docker container và được quản lý bởi một compose stack riêng.

---

## Các dịch vụ có sẵn

| Dịch vụ | URL | Mục đích |
| :--- | :--- | :--- |
| **Caddy Proxy** | — | Định tuyến traffic cho tất cả domain `.test` |
| **DNSMasq** | — | Phân giải domain `*.test` về local |
| **Mailpit** | `https://mail.govard.test` | Bắt email gửi đi để phát triển |
| **PHPMyAdmin** | `https://pma.govard.test` | Quản lý database MySQL qua web |
| **Portainer** | `https://portainer.govard.test` | Giao diện quản lý Docker container |

---

## Truy cập dịch vụ

### Qua CLI Commands

```bash
govard open mail      # Mở Mailpit
govard open db        # Mở PHPMyAdmin
govard open portainer # Mở Portainer
```

### Qua URL trực tiếp

Mở các URL sau trong trình duyệt (HTTPS với SSL tự động):

| Dịch vụ | URL |
| :--- | :--- |
| Mailpit | `https://mail.govard.test` |
| PHPMyAdmin | `https://pma.govard.test` |
| Portainer | `https://portainer.govard.test` |

---

## Thông tin đăng nhập

### Portainer

| Trường | Giá trị |
| :--- | :--- |
| URL | `https://portainer.govard.test` |
| Username | `admin` |
| Password | `AdminGovard123$` |

### PHPMyAdmin

PHPMyAdmin sử dụng cùng database credentials với dự án của bạn. Kết nối bằng:

- **Host**: `mysql` (từ bên trong Docker network)
- **Host**: `127.0.0.1` (từ host, xem `govard db info`)
- **Username/Password**: Từ `.govard.yml` hoặc `env.php` của dự án

### Mailpit

Không cần đăng nhập. Tất cả email gửi đi từ container của dự án sẽ được bắt lại và xem tại `https://mail.govard.test`.

#### Sử dụng Mailpit như SMTP Server

Mailpit hoạt động như một SMTP server, bắt tất cả email gửi đi thay vì gửi đến người nhận thật. Rất hữu ích cho việc phát triển.

| Cài đặt | Giá trị |
| :--- | :--- |
| SMTP Host | `mail` (từ container) hoặc `mail.govard.test` (từ host) |
| SMTP Port | `1025` |
| Username | (để trống) |
| Password | (để trống) |
| SSL/TLS | Tắt |

**Sự khác biệt khi truy cập:**
- **Từ PHP container**: Dùng `mail` (Docker internal DNS resolve container name)
- **Từ máy host**: Dùng `mail.govard.test` (cần DNS resolution qua dnsmasq)

#### Cấu hình Magento 2

**Cách 1: Cấu hình qua `app/etc/env.php`**

Thêm hoặc sửa section `system`:

```php
'system' => [
    'default' => [
        'system' => [
            'smtp' => [
                'host' => 'mail',
                'port' => '1025',
            ],
        ],
    ],
],
```

Hoặc nếu sử dụng module SMTP (như Mageplaza SMTP, Aheadworks SMTP, v.v.):

**Cách 2: Cấu hình trong Admin**

Hầu hết các module SMTP có cài đặt tại **Stores → Configuration → General → SMTP**:

| Trường | Giá trị |
| :--- | :--- |
| Host | `mail` |
| Port | `1025` |
| Username | (để trống) |
| Password | (để trống) |
| SSL/TLS | Tắt |

#### Cấu hình Laravel

Trong `.env`:

```env
MAIL_MAILER=smtp
MAIL_HOST=mail
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME="${APP_NAME}"
```

#### Cấu hình Symfony

Trong `.env`:

```env
MAILER_DSN=smtp://mail:1025
```

Hoặc trong `config/packages/mailer.yaml`:

```yaml
framework:
    mailer:
        dsn: 'smtp://mail:1025'
```

#### Kiểm tra kết nối SMTP

```bash
# Test với swaks
swaks --to test@example.com --server mail --port 1025

# Test với telnet
telnet mail 1025
```

Sau khi gửi thử, truy cập `https://mail.govard.test` để xem email đã bắt được.

---

## Quản lý dịch vụ toàn cục

### Khởi động/Dừng dịch vụ

```bash
# Khởi động tất cả dịch vụ toàn cục
govard svc up

# Dừng tất cả dịch vụ toàn cục
govard svc down

# Khởi động lại tất cả dịch vụ
govard svc restart
```

### Xem trạng thái

```bash
govard svc ps
```

### Xem logs

```bash
govard svc logs
govard svc logs --tail 50
govard svc logs mail
```

### Workflow Sleep/Wake

Tạm dừng tất cả dự án đang chạy để giải phóng tài nguyên:

```bash
govard svc sleep   # Dừng tất cả container của dự án
govard svc wake    # Tiếp tục các dự án đã tạm dừng
```

---

## Tùy chọn nâng cao

### Cờ khởi động

```bash
# Pull image mới nhất trước khi khởi động
govard svc up --pull

# Bỏ qua cài đặt Root CA trust
govard svc up --no-trust

# Tắt automatic local build fallback
govard svc up --no-fallback
```

### Truy cập raw Compose commands

`govard svc` proxy đến Docker Compose cho global services stack:

```bash
govard svc pull
govard svc logs -f
govard svc ps
```

---

## Khắc phục sự cố

### Xung đột port

Nếu dịch vụ toàn cục không khởi động được, kiểm tra xung đột port:

```bash
govard doctor
```

Các xung đột thường gặp:
- Port 80/443: Các web server khác (Apache, Nginx)
- Port 53: Các DNS server khác

### Cảnh báo SSL trên trình duyệt

Nếu thấy cảnh báo SSL sau khi khởi động dịch vụ:

```bash
govard doctor trust
```

### Dịch vụ không phản hồi

Kiểm tra trạng thái container:

```bash
govard svc ps
docker ps | grep govard-proxy
```

Xem logs để xác định vấn đề:

```bash
govard svc logs --tail 100
```

---

## Kiến trúc

Dịch vụ toàn cục chạy từ `~/.govard/proxy/docker-compose.yml` với compose project name là `proxy`.

Các dịch vụ được đăng ký qua Caddy routes:
- `mail.govard.test` → `govard-proxy-mail:8025`
- `pma.govard.test` → `govard-proxy-pma:80`
- `portainer.govard.test` → `govard-proxy-portainer:9000`

---

**[← SSL và Tên miền](/vi/workflows/ssl-and-domains)** | **[Cấu hình →](/vi/reference/configuration)**