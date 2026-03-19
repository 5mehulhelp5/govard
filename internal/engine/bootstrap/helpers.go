package bootstrap

import (
	"crypto/md5" //nolint:gosec // MD5 is intentional here: Magento 1 uses salted MD5 for admin passwords
	"encoding/hex"
	"fmt"
	"strings"
)

// md5SaltedHash returns the MD5 hash of (salt + password) as a hex string.
// This matches Magento 1's salted password hashing: md5(salt . password).
func md5SaltedHash(salt, password string) string {
	h := md5.New() //nolint:gosec
	fmt.Fprint(h, salt+password)
	return hex.EncodeToString(h.Sum(nil))
}

// shellEscape returns a single-quoted shell-safe string suitable for use in sh -c scripts.
// Single quotes in the value are escaped using the standard '"'"' trick.
func shellEscape(s string) string {
	escaped := strings.ReplaceAll(s, "'", "'\"'\"'")
	return "'" + escaped + "'"
}
