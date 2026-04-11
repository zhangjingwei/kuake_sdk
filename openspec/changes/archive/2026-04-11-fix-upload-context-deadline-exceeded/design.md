## Why

上传大文件时，用户遇到"context deadline exceeded"错误，导致上传失败。这个问题在并行上传模式下尤其常见。

根本原因：在 `uploadPartsParallel` 函数中，当某个分片上传失败时，工作协程会调用 `cancel()` 取消上下文，这会导致所有其他正在正常工作的上传协程也被强制终止，返回"context deadline exceeded"错误，而非实际的上传错误。

## What Changes

- 修复并行上传中的上下文取消逻辑，避免单个分片失败导致全部上传中断
- 上传协程在遇到错误时应仅退出自身，不应取消整个上传任务
- 主协程应等待所有工作协程完成后，再检查是否有错误并返回

## Capabilities

### New Capabilities
- none

### Modified Capabilities
- parallel-upload: 修改并行上传的错误处理逻辑，确保单个分片失败不会影响其他分片的上传

## Impact

Affected files:
- `sdk/file.go`: `uploadPartsParallel` 函数 - 修改上下文取消逻辑
- `sdk/file.go`: `upPart` 函数 - 可能需要调整错误处理

Affected behavior:
- 并行上传时，单个分片的瞬时网络错误不应导致整个上传任务取消
- 错误信息将更准确地反映实际失败的分片，而非模糊的"context deadline exceeded"

## Implementation Status

✅ **COMPLETED**

### Changes Made
1. **移除工作协程中的 `ctx.Err()` 检查** (sdk/file.go:235-237, 240-242)
   - 原逻辑：工作协程检查 `ctx.Err()` 后提前退出
   - 新逻辑：工作协程仅在分片上传失败时退出

2. **移除成功结果发送时的 `select { case <-ctx.Done() }`** (sdk/file.go:271-279)
   - 原逻辑：使用 select 防止阻塞
   - 新逻辑：直接发送到 resultCh

3. **移除生产者协程中的 `ctx.Err()` 检查** (sdk/file.go:293-294)
   - 原逻辑：检查上下文错误后返回
   - 新逻辑：仅在文件读取错误时返回

4. **主协程错误处理** (sdk/file.go:387)
   - 原逻辑：调用 `cancel()` 取消所有协程
   - 新逻辑：调用 `close(jobCh)` 关闭任务通道

5. **移除未使用的 `ctx` 和 `cancel` 变量** (sdk/file.go:223-224)

### Test Results
- ✅ 编译通过：`go build ./...`
- ✅ 错误信息格式：`"failed to upload part %d (after %d retries): %w"`
- ✅ 断点续传状态正确保存
- ✅ 并行上传性能不受影响
