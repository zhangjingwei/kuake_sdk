# Robustness Refactor - Security Scan Report

**Date:** 2026-04-13  
**Branch:** feat/robustness-refactor  
**Tool:** gosec  
**Status:** Issues Found - Action Required

## Summary

| Metric | Value |
|--------|-------|
| Files Scanned | 17 |
| Lines Scanned | 6894 |
| Total Issues | 40 |

## Critical Issues (BLOCKER)

### G404: Use of Weak Random Number Generator (math/rand)
**Severity:** HIGH  
**Instances:** 4

**Description:** The code uses `math/rand` with time-based seeding instead of `crypto/rand` for security-sensitive operations like generating random tokens.

**Location:**
- `sdk/share.go:51-52` - `GetShareStoken` method
- `sdk/share.go:111-112` - `GetShareList` method  
- `sdk/share.go:160-161` - `SaveShareFile` method
- `sdk/quark_client.go:44` - Token selection logic

**Fix:** Replace `math/rand` with `crypto/rand` for generating random values that affect security tokens.

**Related to design:** This directly addresses VULN-004 in the design document which was supposed to be fixed in `sdk/share.go:51-52`.

---

### G704: SSRF via Taint Analysis
**Severity:** HIGH  
**Location:** `sdk/quark_client.go:507` - `client.Do(req)`

**Description:** Potential Server-Side Request Forgery (SSRF) vulnerability when making HTTP requests. The client may be vulnerable to SSRF if untrusted URLs are passed to the request builder.

**Recommendation:** Validate and sanitize URLs before making requests. Implement URL whitelist checking.

---

## High Severity Issues

### G101: Potential Hardcoded Credentials
**Severity:** HIGH (LOW confidence)  
**Location:** `sdk/constants.go:51, 63`

**Description:** Hardcoded API endpoint paths detected. While these are not credentials in the traditional sense, they could be considered sensitive configuration.

**Note:** These are API endpoint constants and are generally acceptable for public API endpoints.

---

## Medium Severity Issues

### G301: Directory Permissions Too Permissive
**Severity:** MEDIUM  
**Instances:** 2

**Description:** Directory created with 0755 permissions (world-readable).

**Location:**
- `sdk/file.go:2693` - `MkdirAll(dir, 0755)`
- `sdk/file.go:69` - `MkdirAll(stateDir, 0755)`

**Fix:** Use 0750 or less permissive permissions.

---

### G306: File Write Permissions Too Permissive
**Severity:** MEDIUM  
**Instances:** 2

**Description:** File written with 0644 permissions (world-readable).

**Location:**
- `sdk/file.go:93` - `WriteFile(statePath, data, 0644)`
- `sdk/config.go:116` - `WriteFile(resolvedPath, data, 0644)`

**Fix:** Use 0600 (owner read/write only) for sensitive files.

---

### G505/G501: Blocklisted Cryptographic Primitives
**Severity:** MEDIUM  
**Location:** `sdk/file.go:6-7`

**Description:** Use of weak cryptographic primitives:
- `crypto/md5` - MD5 is cryptographically broken
- `crypto/sha1` - SHA1 is cryptographically broken

**Location Details:**
- Line 6: `crypto/md5` import
- Line 7: `crypto/sha1` import

**Recommendation:** Consider using SHA-256 for file integrity verification. If compatibility with existing systems requires MD5/SHA1, document the security rationale.

---

## Low Severity Issues

### G104: ErrorsUnhandled
**Severity:** LOW  
**Instances:** 11

**Description:** Errors are being ignored (assigned to blank identifier `_`).

**Location:**
- `sdk/file.go:1474, 1442, 1410, 1272, 1252, 1250, 1231, 1219, 1140, 1076, 1038, 69`
- `cmd/main.go:479`

**Specifics:**
- `file.Seek()` errors are unhandled - could indicate file corruption issues
- `os.MkdirAll()` error unhandled - could indicate permission issues
- `os.Stdout.WriteString()` error unhandled - could indicate output issues

**Recommendation:** Log errors or handle them appropriately rather than ignoring.

---

## Validation Layer Scan Results

The new `sdk/validation/` package scanned clean with no issues:

| File | Status |
|------|--------|
| validator.go | Clean |
| checkers.go | Clean |
| default.go | Clean |
| errors.go | Clean |
| pagination.go | Clean |
| random.go | Clean |
| typeassert.go | Clean |

**Note:** The `random.go` in the validation layer correctly uses `crypto/rand` as designed in the design document.

---

## Compliance Check Against Design Document

| Design Item | Status | Notes |
|-------------|--------|-------|
| VULN-001: Path traversal fix | NOT FOUND | No path traversal vulnerability detected in scan |
| VULN-002: Type assertion panic | NOT FOUND | New `typeassert.go` uses safe type assertions |
| VULN-003: CLI arg type assertion | NOT FOUND | `cmd/main.go` uses `strconv.Atoi` with error handling |
| VULN-004: Secure random number | **FAIL** | `sdk/share.go` still uses `math/rand` (lines 51, 111, 161) |
| VULN-005: Pagination limit | NOT FOUND | New `pagination.go` has validation |
| VULN-006: Null pointer access | NOT FOUND | New `typeassert.go` has safe accessors |

---

## Recommendations

### Before Merge - MUST FIX:
1. **Replace `math/rand` with `crypto/rand`** in `sdk/share.go` (lines 51, 111, 161)
2. **Consider fixing SSRF** in `sdk/quark_client.go` by validating URLs
3. **Fix file/directory permissions** to be more restrictive (0750/0600)

### Before Merge - SHOULD FIX:
4. **Handle errors from `file.Seek()`** operations in upload logic
5. **Handle `os.Stdout.WriteString()` error** in `cmd/main.go:479`

### For Future Improvement:
6. **Consider replacing MD5/SHA1** with SHA-256 for file integrity
7. **Add input validation** to prevent SSRF in HTTP client

---

## Scan Artifacts

- **Raw Output:** gosec scan completed with 40 issues
- **Critical Issues:** 1 (G404 - weak random number generator)
- **High Issues:** 1 (G704 - SSRF)
- **Medium Issues:** 3 (G301, G306, G501/G505)
- **Low Issues:** 11 (G104 - unhandled errors)
