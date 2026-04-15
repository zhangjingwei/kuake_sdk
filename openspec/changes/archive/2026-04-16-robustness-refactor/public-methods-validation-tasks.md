# 全量公共方法接入统一校验层 — 任务清单

**目标：** `QuarkClient` 的每个导出方法对外部入参均经过 `sdk/validation`（或显式标注为「无外部入参 / 不适用」），与 `specs/parameter-validation` 及 `design.md` 一致。

**范围：** 与 `sdk/quark_methods_b15_test.go` 一致，共 **24** 个 `(*QuarkClient)` 导出方法（不含仅内部使用的 `upload*`、`listByFid`、`wait*` 等）。

**约定：**

- 返回 `(*StandardResponse, error)`：校验失败 → `Success: false` + 规范 `Code` + `error == nil`（与现有 `UploadFile` 一致）。
- 返回 `(T, error)` 且非 `StandardResponse`：校验失败 → `*validation.ValidationError`（`errors.As`）。
- 仅 `error`：校验失败 → 直接返回该 `error`（可为 `*ValidationError`）。

**单任务粒度：** 每项独立完成时间控制在 **约 5 分钟**（单方法或小类辅助函数 + 单测或扩展现有测试）；若超时则再拆子项。

---

## 阶段 0：门禁与矩阵（2 项）

| # | 任务 | 验收 |
|---|------|------|
| 0.1 | 在 `openspec/.../public-methods-validation-tasks.md`（本文件）或 `micro-tasks-tdd.md` 增加指向：24 方法校验矩阵（方法名 / 入参字段 / 使用的校验器 / 备注）。 | 矩阵文档可审阅 |
| 0.2 | 扩展 `TestB15_ExportedMethodsChecklist` 或新增测试：仅注释说明「本表与校验 rollout 同步更新」，避免与实现脱节。 | `go test ./sdk -run B15 -count=1` 通过 |

---

## 阶段 1：无外部入参（1 项）

| # | 任务 | 验收 |
|---|------|------|
| 1.1 | **`GetUserInfo`、`GetCookies`**：在方法注释中明确「无外部输入参数，不执行入参校验」；若团队要求显式记录，可在矩阵中标记 **N/A**。 | 注释 + 矩阵一行 |

---

## 阶段 2：客户端配置 / 转换（2 项）

| # | 任务 | 验收 |
|---|------|------|
| 2.1 | **`SetBaseURL(baseURL string)`**：`strings.TrimSpace` 后 `NonEmpty()`；可选再限制 scheme（`http`/`https`）或拒绝空，与现有 `normalize` 行为兼容。 | 单测：空串 / 仅空白 → 失败形态符合约定 |
| 2.2 | **`ConvertToFileInfo(qf QuarkFileInfo)`**：对导出可见字段做合理校验（如必填字符串非空）或文档说明「由 API 保证」；二选一必须在矩阵写清。 | 单测或注释 + 矩阵 |

---

## 阶段 3：分享模块 `share.go`（10 项）

| # | 任务 | 验收 |
|---|------|------|
| 3.1 | **`GetShareInfo(text string)`**：`NonEmpty(text)`；失败返回 `*ValidationError`。 | `share_test.go` 增加无效输入用例 |
| 3.2 | **`GetShareStoken(pwdID, passcode string)`**：`pwdID` 使用 `ValidFID()` 或项目约定的 pwd_id 规则（与 `GetShareInfo` 提取的格式一致）；`passcode` 允许空时仅校验 pwdID。 | 单测：非法 pwdID |
| 3.3 | **`GetShareList`**：在现有分页与 sort 校验基础上，补 **`stoken` `NonEmpty`**、`pdirFid` 规则（如 `ValidFID` 或允许 `"0"` 的显式分支）。 | 单测：空 stoken |
| 3.4 | **`SaveShareFile`**：校验 **`pwdID`、`stoken` `NonEmpty`**；`toPdirFid` 若可为 `"0"` 则写清分支。 | 单测或最小表驱动 |
| 3.5 | **`CreateShare(filePath, expireDays, ...)`**：`filePath` `NonEmpty` + 远程路径策略（`ValidPathResult` 或与 `GetFileInfo` 一致）；`expireDays` 使用 `InRange` 或白名单 `{0,1,7,30}`（与注释一致）。 | 单测：非法天数 |
| 3.6 | **`GetShareLink(shareID string)`**：`ValidFID(shareID)` 或等价非空规则。 | 单测 |
| 3.7 | **`SetSharePassword(pwdID, passcode string)`**：二者 `NonEmpty`；若 passcode 允许空则文档 + 分支。 | 单测 |
| 3.8 | **`GetMyShareList`**：矩阵确认已覆盖分页 + order；缺项则补 **order 字段枚举**（与现实现一致）。 | 已有测试 + 矩阵 |
| 3.9 | **`GetShareIDByFid(fid string)`**：`ValidFID(fid)`。 | 单测 |
| 3.10 | **`DeleteShare(shareIDs []string)`**：`len>0` 且每项 `ValidFID`（或文档允许空 slice 的含义）。 | 单测 |

---

## 阶段 4：文件模块 `file.go`（10 项）

| # | 任务 | 验收 |
|---|------|------|
| 4.1 | **`UploadFile`**：对照矩阵复核 `filePath`/`destPath`/`policy`；缺项补测。 | `file_test` 覆盖 |
| 4.2 | **`CreateFolder(folderName, pdirFid string)`**：`folderName` `NonEmpty`；`pdirFid` 与根目录约定（`normalizeRootDir` 前校验）。 | 单测 |
| 4.3 | **`Copy(srcPath, destPath string)`**：入口对 **远程路径** 做 `ValidPathResult` 或与本包 `normalizePath` 组合的统一策略（与 `GetFileInfo` 一致）。 | 单测：穿越路径 |
| 4.4 | **`Move`**：同 4.3 策略。 | 单测 |
| 4.5 | **`Rename(oldPath, newName string)`**：`oldPath` 路径安全 + `newName` `NonEmpty`（及非法字符规则若需要）。 | 单测 |
| 4.6 | **`List(dirPath string)`**：对传入 `dirPath` 在分支前做路径/fid 一致性校验（避免仅依赖下游）。 | 单测 |
| 4.7 | **`GetFileInfo(remotePath, ...)`**：远程路径入口校验与 `skipPathConversion` 分支兼容。 | 单测 |
| 4.8 | **`Delete(remotePath string)`**：同远程路径策略。 | 单测 |
| 4.9 | **`GetDownloadURL(fid string)`**：`ValidFID(fid)`。 | 单测 |
| 4.10 | **`DownloadFile`**：对照矩阵复核 `fid`/`destPath`/`fileName`；补 `fid` `ValidFID` 若缺。 | `file_test` |

---

## 阶段 5：收尾（3 项）

| # | 任务 | 验收 |
|---|------|------|
| 5.1 | 更新 **`specs/parameter-validation/spec.md`** 或 **`design.md`**：Implementation Rule 1 与「24 方法矩阵」交叉引用。 | PR 文档一致 |
| 5.2 | 全量 **`go test ./sdk/...`**。 | 全部通过 |
| 5.3 | **（可选）** 包级 API：`NewQuarkClient`、`LoadConfig`/`SaveConfig` 路径是否纳入「公共入参」— 若纳入，各增 1 个 ≤5min 任务（`configPath` `NonEmpty` 等）。 | 与团队范围一致 |

---

## 24 方法校验矩阵（与 `quark_methods_b15_test.go` / 实现同步维护）

| 方法 | 主要入参 | 校验策略 | 备注 |
|------|----------|----------|------|
| SetBaseURL | baseURL | `TrimSpace` + `NonEmpty`；`url.Parse` 且 scheme 为 `http`/`https` | 无效则**不修改** `baseURL`（保持无 `error` 返回） |
| GetCookies | — | 无外部入参 | N/A |
| ConvertToFileInfo | qf | `NonEmpty`：`Fid`、`Name` | 无效返回 `nil` |
| GetUserInfo | — | 无外部入参 | N/A |
| GetShareInfo | text | `NonEmpty(text)` | `*ValidationError` |
| GetShareStoken | pwdID, passcode | `ValidFID(pwdID)`；passcode 任意 | |
| GetShareList | pwdID, stoken, pdirFid, page, size, sortBy, sortOrder | 分页 `PaginateParams`；`ValidFID(pwdID)`；`NonEmpty(stoken)`；`pdirFid` 为 `0` 或 `ValidFID`；sort 枚举 | |
| SaveShareFile | pwdID, stoken, toPdirFid, … | `NonEmpty(pwdID,stoken)`；`toPdirFid`：`0` 或 `ValidFID` | |
| CreateShare | filePath, expireDays | `validateRemotePathOrFid(filePath)`；`expireDays ∈ {0,1,7,30}` | 入口校验先于 `GetFileInfo` |
| GetShareLink | shareID | `ValidFID(shareID)` | |
| SetSharePassword | pwdID, passcode | `NonEmpty` 二者 | |
| GetMyShareList | page, size, orderField, orderType | 分页；`order_field` / `order_type` 枚举 | 与默认 `PageDefaults` 一致 |
| GetShareIDByFid | fid | `ValidFID(fid)` | |
| DeleteShare | shareIDs | `len>0`；每项 `ValidFID` | |
| UploadFile | filePath, destPath, policy | 已有 `NonEmpty`、`ValidPathResult`、policy 默认 | 见 `file_test` |
| CreateFolder | folderName, pdirFid | `NonEmpty(folderName)`；`pdirFid` 根或 `ValidFID` | |
| Copy / Move | srcPath, destPath | `validateRemotePathOrFid`；`dest` 空时沿用语义不校验 | |
| Rename | oldPath, newName | 远程路径 + `NonEmpty(newName)`；名称不含 `/` `\` | |
| List | dirPath | `validateRemotePathOrFid`（含根） | |
| GetFileInfo / Delete | remotePath | `validateRemotePathOrFid`（含根） | |
| GetDownloadURL | fid | `ValidFID(fid)` | |
| DownloadFile | fid, destPath, fileName | `ValidFID(fid)`；本地 `destPath` 沿用现有 `NonEmpty`/`ValidPathResult` | |

---

## 任务数汇总

| 阶段 | 项数 |
|------|------|
| 0 | 2 |
| 1 | 1 |
| 2 | 2 |
| 3 | 10 |
| 4 | 10 |
| 5 | 3 |
| **合计** | **28** |

若将阶段 5.3 拆为 3 项，则约 **31** 项；仍保持每项约 5 分钟量级（复杂方法若超时，从阶段 3/4 再纵向拆分）。

---

## 说明

- **远程路径 vs 本地路径：** `ValidPath`/`ValidPathResult` 适用于 SDK 约定的**网盘远程路径**；本机 `DownloadFile` 的本地路径需沿用当前 `DownloadFile` 中已有本地校验，勿混用远程规则。
- **重复工作：** `CreateShare` / `Copy` 等会调用 `GetFileInfo`，仍应在**入口**做校验以满足「经过校验层」的明确要求（可先薄封装 `validateRemotePath` 复用）。
