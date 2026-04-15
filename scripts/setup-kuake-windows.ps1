# Kuake Windows 环境设置脚本
# 用于将 kuake.exe 添加到 PATH 环境变量

param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"

# 脚本目录
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$DistDir = Join-Path $ProjectRoot "dist"
$Candidates = Get-ChildItem -Path $DistDir -Filter "kuake-v*-windows-amd64.exe" -File -ErrorAction SilentlyContinue
if (-not $Candidates -or $Candidates.Count -eq 0) {
    Write-Host "错误: 在 dist 目录找不到 kuake-v*-windows-amd64.exe" -ForegroundColor Red
    Write-Host "  目录: $DistDir" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "请先运行仓库根目录的 .\build.sh 或从 Releases 下载 Windows 构建产物" -ForegroundColor Yellow
    exit 1
}
$ExeFile = ($Candidates | Sort-Object LastWriteTime -Descending | Select-Object -First 1).FullName
$BinDir = Join-Path $env:USERPROFILE "bin"
$TargetExe = Join-Path $BinDir "kuake.exe"

Write-Host "=== Kuake Windows 环境设置脚本 ===" -ForegroundColor Green
Write-Host ""

Write-Host "将安装: $ExeFile" -ForegroundColor Cyan
Write-Host ""

# 创建 bin 目录
Write-Host "创建工具目录..." -ForegroundColor Green
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

# 检查是否已存在 kuake.exe
if (Test-Path $TargetExe) {
    if (-not $Force) {
        Write-Host "警告: $TargetExe 已存在" -ForegroundColor Yellow
        $response = Read-Host "是否覆盖? (y/N)"
        if ($response -ne "y" -and $response -ne "Y") {
            Write-Host "已取消操作" -ForegroundColor Yellow
            exit 0
        }
    }
    Write-Host "删除旧文件..." -ForegroundColor Yellow
    Remove-Item -Path $TargetExe -Force
}

# 复制文件
Write-Host "复制可执行文件..." -ForegroundColor Green
Copy-Item -Path $ExeFile -Destination $TargetExe -Force

# 检查 PATH 环境变量
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathArray = $currentPath -split ';' | Where-Object { $_ -ne '' }

if ($pathArray -notcontains $BinDir) {
    Write-Host "添加到 PATH 环境变量..." -ForegroundColor Green
    
    # 添加到用户 PATH
    $newPath = $currentPath
    if ($newPath -and -not $newPath.EndsWith(';')) {
        $newPath += ';'
    }
    $newPath += $BinDir
    
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    
    # 更新当前会话的 PATH
    $env:Path += ";$BinDir"
    
    Write-Host "✓ 已添加到 PATH" -ForegroundColor Green
} else {
    Write-Host "✓ PATH 中已包含 $BinDir" -ForegroundColor Green
}

Write-Host ""
Write-Host "✓ 设置完成!" -ForegroundColor Green
Write-Host ""
Write-Host "文件位置: " -NoNewline
Write-Host $TargetExe -ForegroundColor Cyan
Write-Host ""

# 验证安装
Write-Host "验证安装..." -ForegroundColor Green
try {
    $kuakeCmd = Get-Command kuake -ErrorAction Stop
    Write-Host "✓ kuake 命令可用" -ForegroundColor Green
    Write-Host "  位置: $($kuakeCmd.Source)" -ForegroundColor Cyan
    
    # 测试版本命令
    Write-Host ""
    Write-Host "测试版本命令..." -ForegroundColor Green
    $version = & kuake version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ 版本命令正常" -ForegroundColor Green
        Write-Host $version -ForegroundColor Cyan
    } else {
        Write-Host "⚠ 版本命令返回错误，但文件已安装" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠ kuake 命令在当前会话中不可用" -ForegroundColor Yellow
    Write-Host "  请关闭并重新打开 PowerShell 窗口，然后运行: kuake version" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "下一步:" -ForegroundColor Green
Write-Host "  1. 关闭并重新打开 PowerShell 窗口（使 PATH 生效）" -ForegroundColor Yellow
Write-Host "  2. 运行: " -NoNewline
Write-Host "kuake version" -ForegroundColor Cyan
Write-Host "  3. OpenClaw 集成见仓库 " -NoNewline
Write-Host "openclaw/kuake_skill/" -ForegroundColor Cyan
Write-Host ""
