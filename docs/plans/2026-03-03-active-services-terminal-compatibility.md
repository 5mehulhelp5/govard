I'm using the writing-plans skill to create the implementation plan.

# Active Services Terminal Compatibility Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Sửa lỗi Terminal trong khu vực Active Services (đang fail với Nginx/Redis), đảm bảo các service khác không lặp lại cùng lỗi, bổ sung option để terminate session hiện tại rồi tạo session mới nhanh khi terminal bị treo/lỗi, và cải thiện UX cho các nút Open Remote bằng loading + disable trong lúc action đang chạy.

**Architecture:** Chuẩn hóa lựa chọn shell theo hướng tương thích rộng: default về `sh` (thay vì `bash`) khi UI không truyền shell cụ thể, vì các container Alpine (Nginx/Redis) không có `bash`. Đồng bộ frontend/backend để luôn gửi shell an toàn, bổ sung test cho mapping service target trong Active Services để tránh lệch target ở các service khác. Thêm API lifecycle terminal (terminate session) + UI action “Restart Session” để người dùng chủ động reset kết nối terminal mà không cần reload màn hình. Với Remote actions, ưu tiên bridge-native flow cho `Open Database` và `Open SFTP`; `Open SSH` ưu tiên bridge mở native terminal nếu khả dụng, fallback xterm nếu môi trường không hỗ trợ, đồng thời toàn bộ nút Open phải có trạng thái loading/disabled khi đang xử lý.

**Tech Stack:** Go 1.24 (Wails backend), vanilla JS frontend, Node test runner (`node --test`), Go tests trong `tests/`.

---

### Task 1: Khóa hành vi mong muốn bằng test backend (failing-first)

**Files:**
- Modify: `internal/desktop/test_helpers.go`
- Create: `tests/desktop_shell_normalize_test.go`
- Related: `tests/desktop_shell_target_test.go`

**Step 1: Thêm test wrapper cho normalize shell**

Thêm wrapper theo convention `ForTest`:

```go
func NormalizeShellForTest(shell string) string {
	return normalizeShell(shell)
}
```

**Step 2: Viết test fail cho default shell**

Thêm case trong `tests/desktop_shell_normalize_test.go`:

```go
func TestDesktopPkgNormalizeShellForTest_DefaultsToSh(t *testing.T) {
	if got := desktop.NormalizeShellForTest(""); got != "sh" {
		t.Fatalf("expected sh, got %q", got)
	}
}
```

**Step 3: Viết test fail cho whitelist shell hợp lệ**

```go
func TestDesktopPkgNormalizeShellForTest_RespectsKnownShells(t *testing.T) {
	if got := desktop.NormalizeShellForTest("bash"); got != "bash" { ... }
	if got := desktop.NormalizeShellForTest("sh"); got != "sh" { ... }
}
```

**Step 4: Chạy test để xác nhận đang fail trước khi sửa code**

Run: `go test ./tests/... -run TestDesktopPkgNormalizeShellForTest -v`
Expected: FAIL ở case default shell hiện đang là `bash`.

**Step 5: Commit checkpoint test-first**

```bash
git add internal/desktop/test_helpers.go tests/desktop_shell_normalize_test.go
git commit -m "test(desktop): add shell normalization expectations for service terminal compatibility"
```

---

### Task 2: Sửa backend normalize shell để tương thích cross-service

**Files:**
- Modify: `internal/desktop/actions.go`
- Related: `internal/desktop/terminal.go`

**Step 1: Cập nhật normalize shell**

Đổi logic normalize:

```go
func normalizeShell(shell string) string {
	normalized := strings.ToLower(strings.TrimSpace(shell))
	switch normalized {
	case "bash", "sh":
		return normalized
	default:
		return "sh"
	}
}
```

**Step 2: Giữ nguyên flow StartTerminal, chỉ thay đầu vào shell**

Không đổi contract `StartTerminal(project, service, user, shell)` để tránh ảnh hưởng bridge/UI cũ.

**Step 3: Chạy test backend vừa thêm**

Run: `go test ./tests/... -run TestDesktopPkgNormalizeShellForTest -v`
Expected: PASS.

**Step 4: Chạy regression test liên quan shell target hiện có**

Run: `go test ./tests/... -run TestDesktopPkgResolveShellServiceNameForTest -v`
Expected: PASS (không regress mapping target hiện tại).

**Step 5: Commit backend fix**

```bash
git add internal/desktop/actions.go
git commit -m "fix(desktop): default terminal shell to sh for broader container compatibility"
```

---

### Task 3: Đồng bộ frontend shell defaults với backend

**Files:**
- Modify: `desktop/frontend/modules/logs.js`
- Modify: `desktop/frontend/modules/shell.js`
- Optional touchpoint: `desktop/frontend/modules/terminal.js`

**Step 1: Đổi shell mặc định trong UI từ bash sang sh/auto-safe**

Trong `logs.js`, đổi option ẩn:

```html
<select id="shellCommand" class="hidden">
  <option value="sh">sh</option>
</select>
```

**Step 2: Tránh hardcode fallback bash ở shell controller**

Trong `shell.js`:

```js
const shell = refs.shellCommand?.value || "sh";
```

**Step 3: Cập nhật nhãn terminal cho đúng thực tế**

Đổi text từ `Terminal — bash` thành `Terminal — sh` hoặc `Terminal` (khuyến nghị `Terminal` để tránh hardcode shell).

**Step 4: Chạy frontend test hiện có để bắt regress markup/action**

Run: `node --test tests/frontend/logs_core.test.mjs tests/frontend/dashboard_core.test.mjs`
Expected: PASS.

**Step 5: Commit frontend sync**

```bash
git add desktop/frontend/modules/logs.js desktop/frontend/modules/shell.js
git commit -m "fix(desktop-frontend): align embedded terminal default shell with backend"
```

---

### Task 4: Thêm test cho Active Services target mapping (bao phủ service khác)

**Files:**
- Create: `tests/frontend/dashboard_terminal_targets.test.mjs`
- Related: `desktop/frontend/modules/dashboard.js`

**Step 1: Viết test render service cards với nhiều service name**

Dùng `renderActiveServices(container, env)` với fixture gồm:
- `Nginx` -> kỳ vọng `data-service="web"`
- `Redis` -> kỳ vọng `data-service="redis"`
- `Valkey` -> kỳ vọng `data-service="valkey"`
- `OpenSearch` -> kỳ vọng `data-service="opensearch"`

**Step 2: Xác nhận nút terminal dùng target canonical**

Assert HTML chứa `data-action="open-service-shell"` đi cùng các `data-service` canonical phía trên.

**Step 3: Chạy test mới và xác nhận pass**

Run: `node --test tests/frontend/dashboard_terminal_targets.test.mjs`
Expected: PASS.

**Step 4: Chạy full frontend tests nhanh**

Run: `node --test tests/frontend/*.test.mjs`
Expected: PASS toàn bộ.

**Step 5: Commit test coverage**

```bash
git add tests/frontend/dashboard_terminal_targets.test.mjs
git commit -m "test(frontend): cover active services terminal target mapping"
```

---

### Task 5: Thêm option terminate session hiện tại và tạo session mới

**Files:**
- Modify: `internal/desktop/terminal.go`
- Modify: `internal/desktop/bridge_proxies.go`
- Modify: `desktop/frontend/services/bridge.js`
- Modify: `desktop/frontend/modules/terminal.js`
- Modify: `desktop/frontend/modules/shell.js`
- Modify: `desktop/frontend/modules/logs.js`
- Modify: `desktop/frontend/main.js`
- Modify: `internal/desktop/test_helpers.go`
- Create: `tests/desktop_terminal_session_lifecycle_test.go`
- Modify: `tests/frontend/logs_core.test.mjs`

**Step 1: Thiết kế API terminate session ở backend**

Thêm method mới trên `LogService`:

```go
func (s *LogService) TerminateTerminal(sessionID string) (string, error)
```

Hành vi:
- Nếu session tồn tại: `cancel()`, `pty.Close()`, remove khỏi `sessions` map, emit `terminal:exit`.
- Nếu session không tồn tại: trả message idempotent (không fail hard) để UI vẫn có thể “restart” an toàn.

**Step 2: Expose method qua bridge proxy + frontend bridge service**

- `internal/desktop/bridge_proxies.go`: thêm `TerminateTerminal` proxy.
- `desktop/frontend/services/bridge.js`: thêm `terminateTerminal(id)`.

**Step 3: Thêm action “Restart Session” trong UI terminal**

Trong header terminal (logs panel), thêm nút:

```html
data-action="restart-terminal-session"
```

Nút này phải luôn visible khi có terminal panel, không phụ thuộc hover.

**Step 4: Implement flow restart ở controller**

Trong `shell.js` và `terminal.js`:
- Lưu `currentSessionID` hiện tại.
- `restartSession()` thực hiện:
1. `terminateTerminal(currentSessionID)` (nếu có)
2. gọi lại `startTerminal(...)` với selection hiện tại (project/service/user/shell)
3. resize lại terminal.

**Step 5: Hook action ở main event handler**

Trong `main.js`, bắt `data-action="restart-terminal-session"` và gọi controller tương ứng.

**Step 6: Thêm test backend lifecycle**

Trong `tests/desktop_terminal_session_lifecycle_test.go`, test các case:
- terminate session không tồn tại -> không panic, trả message hợp lệ.
- terminate session tồn tại -> đóng session và remove khỏi store.

Sử dụng wrapper `ForTest` nhỏ trong `internal/desktop/test_helpers.go` nếu cần để truy cập trạng thái session mà không export rộng.

**Step 7: Thêm test frontend contract cho restart action**

Update `tests/frontend/logs_core.test.mjs` assert có:
- `data-action="restart-terminal-session"`.

**Step 8: Chạy test scope task này**

Run:
- `go test ./tests/... -run TestDesktopPkgTerminalSessionLifecycle -v`
- `node --test tests/frontend/logs_core.test.mjs`

Expected: PASS.

**Step 9: Commit terminal restart feature**

```bash
git add internal/desktop/terminal.go internal/desktop/bridge_proxies.go desktop/frontend/services/bridge.js desktop/frontend/modules/terminal.js desktop/frontend/modules/shell.js desktop/frontend/modules/logs.js desktop/frontend/main.js internal/desktop/test_helpers.go tests/desktop_terminal_session_lifecycle_test.go tests/frontend/logs_core.test.mjs
git commit -m "feat(desktop): add terminal session restart option via terminate-and-reconnect flow"
```

---

### Task 6: Chuyển Remote Open buttons sang bridge-first và thêm loading/disable state

**Files:**
- Modify: `internal/desktop/remotes.go`
- Modify: `internal/desktop/bridge_proxies.go`
- Modify: `desktop/frontend/services/bridge.js`
- Modify: `desktop/frontend/modules/shell.js`
- Modify: `desktop/frontend/modules/remotes.js`
- Modify: `desktop/frontend/state/store.js`
- Modify: `desktop/frontend/main.js`
- Modify: `tests/frontend/remotes_core.test.mjs`
- Optional: `tests/desktop_remote_actions_test.go` (nếu cần phủ backend flow mới)

**Step 1: Thêm bridge methods cho remote open actions**

Thêm API backend cho UI gọi trực tiếp thay vì luôn đi qua xterm:
- `OpenRemoteDB(project, remoteName)`
- `OpenRemoteSFTP(project, remoteName)`
- `OpenRemoteShell(project, remoteName)` (ưu tiên native terminal nếu khả dụng, fallback xterm)

Lưu ý:
- Reuse logic `open db -e <remote> --client` và `open sftp -e <remote>` hiện có để tránh duplicate business logic.
- Method trả message rõ ràng cho UI (`opened`, `unsupported`, `fallback-to-terminal`).

**Step 2: Expose qua Wails bridge + JS bridge client**

- `internal/desktop/bridge_proxies.go`: thêm các proxy tương ứng.
- `desktop/frontend/services/bridge.js`: thêm `openRemoteDB`, `openRemoteSFTP`, `openRemoteShell`.

**Step 3: Thêm state pending cho từng remote open action**

Trong `state/store.js`, thêm state dạng map:

```js
remoteActionPending: {}
```

Key đề xuất: `${remoteName}:${action}` với action thuộc `ssh|db|sftp`.

**Step 4: Render loading indicator + disabled cho 3 nút Open**

Trong `remotes.js`:
- Khi pending true:
  - thêm `disabled aria-disabled=\"true\"`
  - đổi icon/text sang loading (ví dụ spinner + `Opening...`)
  - giảm opacity và cursor `not-allowed`
- Khi pending false: render trạng thái bình thường.

**Step 5: Hook action handlers để set/unset pending an toàn**

Trong `main.js` hoặc controller liên quan:
1. set pending trước khi gọi bridge.
2. gọi API bridge tương ứng.
3. unset pending trong `finally` (đảm bảo dù fail vẫn enable lại).

Anti-double-click:
- Nếu pending đã true thì ignore click tiếp theo.

**Step 6: Điều chỉnh shell controller usage**

- `openRemoteDB/openRemoteSFTP` không còn phụ thuộc `startGovardTerminal` qua xterm.
- `openRemoteShell`:
  - bridge-native path trước;
  - fallback xterm path khi backend trả trạng thái fallback.

**Step 7: Bổ sung frontend tests**

Update `tests/frontend/remotes_core.test.mjs`:
- Assert nút Open hỗ trợ trạng thái disabled/loading markup.
- Assert action attributes giữ nguyên (`open-remote-shell`, `open-remote-db`, `open-remote-sftp`).

**Step 8: Chạy test scope task này**

Run:
- `node --test tests/frontend/remotes_core.test.mjs`
- `node --test tests/frontend/*.test.mjs`

Expected: PASS.

**Step 9: Commit remote open UX + bridge migration**

```bash
git add internal/desktop/remotes.go internal/desktop/bridge_proxies.go desktop/frontend/services/bridge.js desktop/frontend/modules/shell.js desktop/frontend/modules/remotes.js desktop/frontend/state/store.js desktop/frontend/main.js tests/frontend/remotes_core.test.mjs
git commit -m "feat(desktop-remotes): use bridge-first open actions with loading and disabled button states"
```

---

### Task 7: Verification cuối và checklist release-safe

**Files:**
- N/A (verification)

**Step 1: Format/check Go**

Run: `gofmt -s -w internal/desktop/actions.go internal/desktop/terminal.go internal/desktop/remotes.go internal/desktop/bridge_proxies.go internal/desktop/test_helpers.go tests/desktop_shell_normalize_test.go tests/desktop_terminal_session_lifecycle_test.go`
Expected: không còn diff formatting.

**Step 2: Chạy test scope thay đổi**

Run: `go test ./tests/... -run 'TestDesktopPkgNormalizeShellForTest|TestDesktopPkgResolveShellServiceNameForTest|TestDesktopPkgTerminalSessionLifecycle|TestDesktopRemote' -v`
Expected: PASS.

**Step 3: Chạy frontend tests liên quan**

Run: `node --test tests/frontend/*.test.mjs`
Expected: PASS.

**Step 4: Chạy suite nhanh theo project convention**

Run: `make test-fast`
Expected: PASS.

**Step 5: Manual smoke test desktop (bắt buộc cho bug này)**

Run:
- `DISPLAY=:1 govard desktop --dev`
- Mở `http://localhost:34115`
- Tại Active Services, mở terminal lần lượt cho: `Nginx(web)`, `Redis`, `PHP`, `DB` (nếu có), `Valkey` (nếu có)
- Với mỗi service, bấm `Restart Session` ít nhất 1 lần.
- Ở tab Remotes, bấm `Open SSH`, `Open Database`, `Open SFTP` và xác nhận:
  - nút vào trạng thái loading + disabled ngay khi click
  - click lặp không trigger request trùng
  - sau success/fail, nút tự trở lại trạng thái enabled

Expected:
- terminal mở được và không exit ngay vì `bash not found`.
- `Restart Session` đóng session cũ và mở session mới thành công.
- input command cơ bản chạy được (`whoami`, `pwd`, `ls`).

**Step 6: Final review checklist**

- `git status` chỉ chứa file liên quan.
- Không làm hỏng API bridge hiện có (chỉ add method mới).
- Không làm hỏng log/service filters ở tab Logs.
- `Open Database` và `Open SFTP` ở Remotes không còn phụ thuộc xterm path mặc định.

---

## Risk Notes

- Đổi default shell từ `bash` -> `sh` có thể giảm trải nghiệm interactive cho user thích bash trong PHP container; nhưng đây là tradeoff hợp lý để đảm bảo terminal hoạt động ổn định trên tất cả service (đặc biệt Alpine-based).
- Restart flow cần idempotent: người dùng bấm nhiều lần liên tiếp không được làm crash session map.
- Remote loading/disable state cần luôn reset trong `finally`; nếu không sẽ gây nút bị kẹt disabled sau lỗi.
- Nếu cần giữ `bash` cho một số service, có thể follow-up bằng cơ chế `auto` thông minh (probe shell trong container) sau khi fix nóng này ổn định.

## Rollout Strategy

1. Merge fix shell compatibility + session restart + remote open loading/disable + tests.
2. QA manual với project có `nginx + redis`.
3. QA thêm profile có `valkey/opensearch/varnish` nếu có sẵn.
4. Theo dõi phản hồi desktop users 1-2 ngày, rồi cân nhắc nâng cấp cơ chế shell auto-probe.
