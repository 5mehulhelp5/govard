# Cài đặt

Trang này hướng dẫn tất cả các phương pháp cài đặt Govard trên Linux và macOS.

::: warning QUAN TRỌNG
Không nên trộn lẫn các kênh cài đặt trên cùng một máy (ví dụ: `.deb` + `make install` + `self-update` ở các đường dẫn khác nhau). Chỉ sử dụng **một kênh duy nhất** để tránh xung đột binary trong `/usr/bin` và `/usr/local/bin`.
:::

---

## 🚀 Cài đặt một dòng lệnh (Linux/macOS)

Cài đặt phiên bản release mới nhất bằng một lệnh duy nhất:

```bash
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash
```

Sử dụng `wget`:

```bash
wget -qO- https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash
```

### Các tùy chọn cài đặt phổ biến

```bash
# Cài vào ~/.local/bin (không cần sudo)
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --local

# Build từ source (tự động cài Go 1.25 nếu cần)
curl -fsSL https://raw.githubusercontent.com/ddtcorex/govard/master/install.sh | bash -s -- --source
```

Mặc định, script sẽ cài đặt cả `govard` (CLI) và `govard-desktop` (Desktop app) vào `/usr/local/bin` và:
- Tự động phát hiện và cài đặt các system dependencies còn thiếu (`certutil`, `WebKitGTK`).
- Khởi chạy các global services.
- Cấu hình SSL trust.
- Trên Linux, tự động fallback sang giải nén `govard-desktop` từ package `.deb` nếu file nén archive độc lập không có sẵn trong bản release.

---

## 📦 Release Installers (CLI + Desktop)

Mỗi release được gắn tag đều publish các package cài đặt bao gồm cả `govard` (CLI) và `govard-desktop`.

Tải từ [trang releases](https://github.com/ddtcorex/govard/releases):

### Linux (`.deb`)

```bash
sudo dpkg -i govard_<version>_linux_amd64.deb
```

### macOS (`.pkg`)

```bash
sudo installer -pkg govard_<version>_Darwin_arm64.pkg -target /
```

---

## 🔧 Build từ Source

### Điều kiện tiên quyết

Đảm bảo bạn đã cài đặt các công cụ sau:

| Công cụ | Phiên bản yêu cầu |
| :--- | :--- |
| Go | `1.25+` |
| Node.js | `20+` |
| Yarn | v1.x |
| golangci-lint | v2.11+ |
| Docker + Docker Compose | Bản mới nhất |
| Wails | `v2.11+` (chỉ khi phát triển desktop app) |

### Cài đặt từ Source

```bash
git clone https://github.com/ddtcorex/govard.git
cd govard
./install.sh --source
```

### Thiết lập cho local development

1. **Cài đặt Go 1.25+** từ [go.dev](https://go.dev/dl/)

2. **Kích hoạt Yarn** qua Corepack:
   ```bash
   corepack enable
   ```

3. **Cài đặt golangci-lint**:
   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

4. **Cài đặt Wails** (để phát triển desktop):
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   wails version
   ```

Không cần quyền `sudo` — bạn có thể cài đặt mọi thứ ở local và cập nhật biến `PATH`.

---

## 🐳 Docker Images

Govard sử dụng một Dockerfile PHP duy nhất với build args thay vì các thư mục phân chia theo version.

```bash
# Image PHP tiêu chuẩn
docker build -f docker/php/Dockerfile \
  -t ddtcorex/govard-php:8.4 \
  --build-arg PHP_VERSION=8.4 \
  docker/php

# Image PHP tối ưu hóa riêng cho Magento 2
docker build -f docker/php/magento2/Dockerfile \
  -t ddtcorex/govard-php-magento2:8.4 \
  --build-arg PHP_VERSION=8.4 \
  docker/php
```

---

## 🔄 Cập nhật Govard

```bash
govard self-update
```

Lệnh `self-update` tự động tải về release artifact phù hợp với nền tảng, **xác minh mã băm SHA-256 checksum**, và thay thế các file binary đã cài đặt một cách atomic (`govard` + `govard-desktop`).

---

## ✅ Xác minh cài đặt

```bash
govard version
govard doctor
```

Lệnh `govard doctor` chạy các chẩn đoán hệ thống (system diagnostics) bao gồm kiểm tra Docker, DNS, ports và SSL trust store.

---

**[← Trang chủ](/vi/)** | **[Dự án đầu tiên →](/vi/getting-started/getting-started)**