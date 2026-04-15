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
$ExeFile = Join-Path $DistDir "kuake-v1.4.2-windows-amd64.exe"
$BinDir = Join-Path $env:USERPROFILE "bin"
$TargetExe = Join-Path $BinDir "kuake.exe"

Write-Host "=== Kuake Windows 环境设置脚本 ===" -ForegroundColor Green
Write-Host ""

# 检查可执行文件是否存在
if (-not (Test-Path $ExeFile)) {
    Write-Host "错误: 找不到可执行文件" -ForegroundColor Red
    Write-Host "  预期位置: $ExeFile" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "请确保已编译项目或从 Releases 下载了可执行文件" -ForegroundColor Yellow
    exit 1
}

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
Write-Host "  3. 如果正常，运行部署脚本: " -NoNewline
Write-Host ".\scripts\deploy-openclaw-skill.ps1" -ForegroundColor Cyan
Write-Host ""
