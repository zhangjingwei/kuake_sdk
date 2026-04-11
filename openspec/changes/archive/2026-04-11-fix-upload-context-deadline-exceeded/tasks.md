## 1. 修复 uploadPartsParallel 错误处理逻辑 ✅

- [x] 1.1 移除工作协程中的 `cancel()` 调用（file.go:269 和 file.go:308）
- [x] 1.2 工作协程失败时仅将错误发送到 resultCh，然后直接 return 退出
- [x] 1.3 移除生产者协程中因 ctx.Err() 而返回的逻辑（file.go:294-295, 303-308）
- [x] 1.4 确保主协程在接收到第一个错误后关闭 jobCh 以通知其他协程停止

## 2. 验证错误信息准确性 ✅

- [x] 2.1 确保返回的错误包含实际失败的分片号和错误原因
- [x] 2.2 验证错误格式：`"failed to upload part %d (after %d retries): %w"`

## 3. 测试 ✅

- [x] 3.1 模拟单个分片上传失败，验证错误信息准确
- [x] 3.2 模拟多个分片失败，验证只报告第一个错误
- [x] 3.3 验证已上传的分片状态正确保存（断点续传）
- [x] 3.4 验证并行上传性能不受影响

## 实施细节

### 修改的文件
- `sdk/file.go`: `uploadPartsParallel` 函数

### 主要变更
1. 移除工作协程中的所有 `ctx.Err()` 检查（2处）
2. 移除生产者协程中的 `ctx.Err()` 检查（1处）
3. 移除成功结果发送时的 `select { case <-ctx.Done() }` 检查
4. 主协程接收到错误后调用 `close(jobCh)` 而非 `cancel()`
5. 移除未使用的 `ctx` 和 `cancel` 变量声明

### 错误处理流程
```
工作协程失败 → 发送错误到 resultCh → 主协程设置 firstErr → 主协程关闭 jobCh → 
等待其他协程完成 → 返回 firstErr
```
