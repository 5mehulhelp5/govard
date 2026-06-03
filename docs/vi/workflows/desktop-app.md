# Ứng dụng Desktop (Desktop App)

Govard Desktop là ứng dụng giao diện (GUI) viết bằng Wails, tái sử dụng cùng một core engine với phiên bản CLI.

---

## Các chế độ khởi chạy (Launch Modes)

```bash
govard desktop              # Khởi chạy ứng dụng desktop đã được build
govard desktop --dev        # Chạy ở chế độ phát triển Wails dev (live backend)
govard desktop --background # Khởi động ẩn, tái sử dụng instance khi mở lại
```

| Chế độ | Mô tả |
| :--- | :--- |
| `govard desktop` | Khởi chạy tiêu chuẩn — sử dụng binary đã build |
| `govard desktop --dev` | Chế độ phát triển — live Go backend, hỗ trợ hot reload frontend |
| `govard desktop --background` | Tiến trình chạy ngầm — giữ ứng dụng hoạt động ngay cả khi đóng cửa sổ |

---

## Giao diện hiện tại

Ứng dụng Desktop tập trung vào các thao tác quản lý cốt lõi:

| Tính năng | Mô tả |
| :--- | :--- |
| **Bảng điều khiển môi trường** | Start/stop/open/delete — bao gồm việc phát hiện các dự án Docker mồ côi |
| **Workspace dự án** | Danh sách các môi trường, các thao tác nhanh, quy trình onboard |
| **Thao tác nhanh** | PHPMyAdmin, bật/tắt Xdebug, kiểm tra sức khỏe, Mailpit, DB client |
| **Tab Remotes** | Quy trình thêm/kiểm tra/mở/lập kế hoạch đồng bộ cho môi trường remote |
| **Giám sát tài nguyên** | CPU, RAM, mạng, cảnh báo OOM |
| **Nhật ký (Logs)** | Lọc logs theo dịch vụ, mức độ nghiêm trọng, tìm kiếm text, stream trực tiếp |
| **Shell Launcher** | Chọn container dịch vụ, user kết nối và loại shell tương ứng |
| **Thông báo hệ thống** | Hiển thị cảnh báo khi các thao tác thành công hoặc thất bại |
| **Bảng Cài đặt** | Thay đổi theme giao diện, cấu hình proxy, trình duyệt ưa thích, DB client |

> Các thao tác khởi động/dừng/tải môi trường và quản lý dịch vụ toàn cục trên Desktop đều gọi trực tiếp tới tầng lệnh CLI của Govard (`govard up`, `govard env ...`, `govard svc ...`), giúp hành vi của ứng dụng Desktop luôn đồng bộ với các cập nhật mới nhất của CLI.

---

## Phím tắt (Keyboard Shortcuts)

| Phím tắt | Thao tác |
| :--- | :--- |
| `Ctrl+,` / `Cmd+,` | Mở Cài đặt |
| `Esc` | Đóng Cài đặt |

---

## Các thao tác Remote trên Desktop

| Thao tác | Hành vi |
| :--- | :--- |
| Mở Database (Remote) | Gọi lệnh `govard open db -e <remote> --client` |
| Mở terminal SSH (Remote) | Ưu tiên các terminal gốc của Linux, fallback về giao thức `ssh://` |
| Mở SFTP (Remote) | Ưu tiên ứng dụng FileZilla, fallback về giao thức `sftp://` |

Đối với phương thức cấu hình `auth.method: ssh-agent`, ứng dụng Desktop tái sử dụng `SSH_AUTH_SOCK` và thăm dò socket tại `/run/user/<uid>/keyring/ssh` trên môi trường Linux.

### Mở Cơ sở dữ liệu local

- Xác định host và port docker được publish trước.
- Fallback về PHPMyAdmin nếu trình kết nối DB client được cấu hình bị thất bại.

---

## Cấu hình ưu tiên của Desktop (Desktop Preferences)

Các cài đặt cấu hình ưu tiên được lưu trữ tại file:

```
~/.govard/desktop-preferences.json
```

Các cấu hình ưu tiên hiện được ghi nhớ:
- Theme giao diện (light/dark)
- Proxy target
- Trình duyệt ưa thích
- Trình kết nối database client ưa thích

---

## Chế độ phát triển (Dev Mode)

Khi phát triển ứng dụng Desktop từ source code, bạn cần cấu hình display server để tránh Wails bị crash trong môi trường headless:

```bash
DISPLAY=:1 govard desktop --dev
```

Chế độ Wails dev biên dịch mã nguồn backend và khởi chạy server frontend tại địa chỉ:

```
http://localhost:34115
```

Đây là cách kiểm thử trên trình duyệt được khuyên dùng vì cầu nối Go backend (Go backend bridge) hoạt động trực tiếp để tải dữ liệu dự án thực tế.

---

## Cấu trúc thư mục Frontend

| File | Mục đích |
| :--- | :--- |
| `desktop/frontend/index.html` | Điểm vào HTML chính |
| `desktop/frontend/main.js` | Khởi tạo, lắng nghe sự kiện, quản lý tab và state |
| `desktop/frontend/services/bridge.js` | Cầu nối gọi RPC tới Go backend của Wails |
| `desktop/frontend/state/store.js` | State UI dùng chung (dự án đang chọn, bộ lọc) |
| `desktop/frontend/modules/` | Các module tính năng (dashboard, logs, remotes, v.v.) |
| `desktop/frontend/ui/toast.js` | Hệ thống hiển thị thông báo toast |
| `desktop/frontend/utils/dom.js` | Các helper xử lý DOM dùng chung |

### Hành vi ở chế độ test (Test Mode)

| Cách thức truy cập | Trạng thái Backend | Dữ liệu hiển thị |
| :--- | :--- | :--- |
| Wails dev (`localhost:34115`) | Hoạt động đầy đủ cầu nối backend | Dữ liệu dự án thực tế |
| Mở trực tiếp file HTML (không backend) | Cầu nối không khả dụng | Dữ liệu mock fallback + toast cảnh báo |

---

## Ghi chú về Kiến trúc (Architecture Notes)

Ứng dụng Desktop được thiết kế tập trung tối đa vào các quy trình thao tác vận hành:

- Điểm vào desktop: `cmd/govard-desktop`
- Wails bindings: `internal/desktop`
- Khung giao diện frontend: `desktop/frontend/index.html`
- Khởi tạo & Lắng nghe sự kiện: `desktop/frontend/main.js`
- Cầu nối gọi backend: `desktop/frontend/services/bridge.js`
- Quản lý State: `desktop/frontend/state/store.js`
- Module tính năng: `desktop/frontend/modules/`

Để hiểu sâu hơn về kiến trúc hệ thống, xem thêm tài liệu [Kiến trúc](/vi/developer/architecture).

---

[SSL và Tên miền](/vi/workflows/ssl-and-domains) | [Kiến trúc](/vi/developer/architecture)