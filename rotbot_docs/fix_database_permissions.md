# SQLite数据库权限问题修复指南

## 🚨 问题描述
错误信息：`attempt to write a readonly database (8)`
这表明SQLite数据库文件是只读的，无法进行写入操作。

## 🔍 问题分析
从权限检查结果看，数据库文件权限设置有问题：
- 当前用户只有 `ReadAndExecute` 权限
- 缺少 `Write` 和 `Modify` 权限
- 导致配额更新等写操作失败

## 🛠️ 解决方案

### 方案1：修复文件权限（推荐）

#### Windows PowerShell命令：
```powershell
# 1. 停止正在运行的服务
# 按 Ctrl+C 停止当前运行的服务

# 2. 给当前用户添加完全控制权限
icacls "data\gateway.db" /grant "%USERNAME%:(F)"

# 3. 给data目录也添加权限
icacls "data" /grant "%USERNAME%:(F)"

# 4. 验证权限设置
Get-Acl data\gateway.db | Format-List
```

#### 或者使用图形界面：
1. 右键点击 `data\gateway.db` 文件
2. 选择"属性" → "安全"选项卡
3. 点击"编辑"按钮
4. 选择你的用户名
5. 勾选"完全控制"权限
6. 点击"确定"保存

### 方案2：重新创建数据库

如果权限修复不起作用，可以重新创建数据库：

```powershell
# 1. 停止服务
# 按 Ctrl+C 停止当前运行的服务

# 2. 备份现有数据库（可选）
Copy-Item "data\gateway.db" "data\gateway.db.backup"

# 3. 删除现有数据库
Remove-Item "data\gateway.db"

# 4. 重新运行迁移创建数据库
go run cmd/migrate/main.go -direction=up

# 5. 重新设置测试数据
go run cmd/e2etest/main.go -action=setup
```

### 方案3：使用不同的数据库路径

修改配置文件使用用户目录下的数据库：

```yaml
# configs/config.yaml
database:
  driver: "sqlite"
  dsn: "%USERPROFILE%/ai-gateway/gateway.db"  # 使用用户目录
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s
```

然后创建目录并重新初始化：
```powershell
# 创建目录
New-Item -ItemType Directory -Path "$env:USERPROFILE\ai-gateway" -Force

# 运行迁移
go run cmd/migrate/main.go -direction=up

# 设置测试数据
go run cmd/e2etest/main.go -action=setup
```

## 🔧 快速修复脚本

创建一个PowerShell脚本来自动修复：

```powershell
# fix_permissions.ps1
Write-Host "🔧 修复AI API Gateway数据库权限问题..." -ForegroundColor Green

# 检查数据库文件是否存在
if (Test-Path "data\gateway.db") {
    Write-Host "📁 找到数据库文件，正在修复权限..." -ForegroundColor Yellow
    
    # 修复文件权限
    icacls "data\gateway.db" /grant "$env:USERNAME:(F)" /T
    icacls "data" /grant "$env:USERNAME:(F)" /T
    
    Write-Host "✅ 权限修复完成！" -ForegroundColor Green
    
    # 验证权限
    Write-Host "🔍 验证权限设置：" -ForegroundColor Cyan
    Get-Acl data\gateway.db | Select-Object Owner, Access | Format-List
} else {
    Write-Host "❌ 数据库文件不存在，需要重新创建" -ForegroundColor Red
    Write-Host "请运行: go run cmd/migrate/main.go -direction=up" -ForegroundColor Yellow
}

Write-Host "🚀 现在可以重新启动服务了！" -ForegroundColor Green
```

## 🚀 重启服务

权限修复完成后，重新启动服务：

```powershell
# 启动服务
go run cmd/server/main.go -config configs/config.yaml
```

## 🧪 验证修复

运行测试确认问题已解决：

```powershell
# 运行健康检查
curl http://localhost:8080/health

# 运行完整测试
python test_service.py

# 测试API调用
curl -X POST http://localhost:8080/v1/chat/completions ^
  -H "Authorization: Bearer ak_4278fc65b1d32cc99fe69fc25bf352261fab3aa0b08488d919dce0097b0f3915" ^
  -H "Content-Type: application/json" ^
  -d "{\"model\":\"gpt-3.5-turbo\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}"
```

## 🔍 预防措施

为避免类似问题再次发生：

1. **使用专门的数据目录**：
   ```yaml
   database:
     dsn: "C:/ai-gateway-data/gateway.db"
   ```

2. **设置正确的目录权限**：
   ```powershell
   New-Item -ItemType Directory -Path "C:\ai-gateway-data" -Force
   icacls "C:\ai-gateway-data" /grant "$env:USERNAME:(F)" /T
   ```

3. **使用Docker部署**（推荐）：
   ```bash
   docker-compose up -d
   ```
   Docker会自动处理权限问题。

## 📝 注意事项

- 在生产环境中，建议使用PostgreSQL而不是SQLite
- 确保数据库文件和目录都有正确的权限
- 定期备份数据库文件
- 考虑使用专门的数据库用户和权限设置
