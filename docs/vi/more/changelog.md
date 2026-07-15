---
title: Nhật ký thay đổi Govard
description: Lịch sử phiên bản đầy đủ của Govard theo Semantic Versioning, chi tiết tính năng mới, sửa lỗi và thay đổi blueprint.
---

# Nhật ký thay đổi (Changelog)

Govard tuân thủ quy tắc [Semantic Versioning](https://semver.org/spec/v2.0.0.html) và duy trì nhật ký thay đổi chi tiết trực tiếp trong kho lưu trữ mã nguồn.

---

## Nhật ký thay đổi đầy đủ

Lịch sử các phiên bản phát hành đầy đủ được duy trì tại:

**[CHANGELOG.md trên GitHub](https://github.com/ddtcorex/govard/blob/master/CHANGELOG.md)**

---

## Các bản phát hành mới nhất

Để xem ghi chú chi tiết và các liên kết tải về cho các phiên bản phát hành mới nhất, truy cập địa chỉ:

**[GitHub Releases Page](https://github.com/ddtcorex/govard/releases)**

---

## Các Artifact phát hành (Release Artifacts)

Mỗi bản phát hành được tag (`vX.Y.Z`) sẽ publish:

| Artifact | Mô tả |
| :--- | :--- |
| `govard_<version>_Linux_amd64.tar.gz` | Lưu trữ binary CLI (Linux x86_64) |
| `govard_<version>_Linux_arm64.tar.gz` | Lưu trữ binary CLI (Linux ARM64) |
| `govard_<version>_Darwin_amd64.tar.gz` | Lưu trữ binary CLI (macOS Intel) |
| `govard_<version>_Darwin_arm64.tar.gz` | Lưu trữ binary CLI (macOS Apple Silicon) |
| `govard_<version>_linux_amd64.deb` | Gói cài đặt Linux (CLI + Desktop) |
| `govard_<version>_Darwin_arm64.pkg` | Gói cài đặt macOS (CLI + Desktop) |
| `checksums.txt` | Mã băm SHA-256 xác minh cho mọi artifact |

---

## Luôn cập nhật phiên bản mới nhất

```bash
# Kiểm tra phiên bản hiện tại
govard version

# Cập nhật lên bản mới nhất
govard self-update
```

Lệnh `govard self-update` sẽ tự động tải về artifact phát hành tương ứng với hệ điều hành của bạn, xác minh mã băm SHA-256 và thực hiện thay thế binary một cách an toàn.

---

[FAQ](/vi/more/faq) | [Bắt đầu](/vi/getting-started/getting-started)