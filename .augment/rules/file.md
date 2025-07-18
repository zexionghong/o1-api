---
type: "agent_requested"
description: "告知ai，要复用项目的组件或者工具，不要每次都生成一个，尽量使用现在有的依赖来实现功能"
---
# 开发规范与最佳实践

## 🎯 核心原则

### 1. 复用优先原则 (Reuse First)
- **优先使用现有组件**：在开发新功能前，必须先检查项目中是否已有类似的组件或工具
- **避免重复造轮子**：不要为了相同或相似的功能创建新的组件
- **扩展而非重写**：如果现有组件接近需求但不完全匹配，优先考虑扩展现有组件

### 2. 依赖管理原则
- **使用现有依赖**：优先使用项目中已安装的依赖包来实现功能
- **谨慎添加新依赖**：添加新依赖前必须评估：
  - 是否有现有依赖可以实现相同功能
  - 新依赖的维护状态和社区活跃度
  - 对项目包大小和性能的影响
- **依赖版本管理**：保持依赖版本的一致性和稳定性

### 3. 文档驱动开发
- **查阅现有文档**：开发前必须阅读相关的项目文档和API文档
- **记录所有变更**：每次功能开发或修改都必须更新相应文档
- **保持文档同步**：代码变更时同步更新文档，确保文档的准确性

## 📋 开发流程规范

### 开发前检查清单
- [ ] 检查现有组件库是否有可复用的组件
- [ ] 查看项目依赖列表，确认是否有合适的工具库
- [ ] 阅读相关功能的现有文档和API规范
- [ ] 检查是否有类似功能的实现可以参考

### 开发中规范
- [ ] 优先扩展现有组件而非创建新组件
- [ ] 使用项目统一的代码风格和命名规范
- [ ] 遵循项目的架构模式和设计原则
- [ ] 编写清晰的代码注释和文档字符串

### 开发后规范
- [ ] 完成一个小功能后都需要写单元测试，验证功能可行性
- [ ] 更新相关的API文档，文档放置在rotbot_docs目录下，没有的话创建一个
- [ ] 更新功能说明文档
- [ ] 记录变更日志
- [ ] 更新使用示例和教程

## 🔧 具体实施指南

### 前端开发
```typescript
// ✅ 好的做法：复用现有组件
import { ToolCreateDialog } from '../tool-create-dialog';
import { ToolEditDialog } from '../tool-edit-dialog';

// ❌ 避免：为相似功能创建新组件
// import { ToolUpdateDialog } from '../tool-update-dialog';
```

### 后端开发
```go
// ✅ 好的做法：复用现有服务和仓储
func (s *ToolService) UpdateUserToolInstance(ctx context.Context, id string, userID int64, req *entities.UpdateUserToolInstanceRequest) (*entities.UserToolInstance, error) {
    // 复用现有的验证逻辑
    instance, err := s.toolRepo.GetUserToolInstanceByID(ctx, id)
    // ...
}

// ❌ 避免：重复实现相同的验证逻辑
```

### 依赖使用示例
```json
// package.json - 优先使用现有依赖
{
  "dependencies": {
    "@mui/material": "^5.x.x",  // 已有UI库，不要添加其他UI库
    "react-i18next": "^13.x.x", // 已有国际化，不要添加其他i18n库
    "axios": "^1.x.x"           // 已有HTTP客户端，不要添加fetch库
  }
}
```

## 📚 文档更新规范

### 必须更新的文档类型
1. **API文档** (`docs/API.md`)
   - 新增或修改的API接口
   - 请求/响应格式变更
   - 错误码和状态码

2. **功能文档** (`README.md`, `USAGE_GUIDE.md`)
   - 新功能的使用说明
   - 配置参数的变更
   - 操作流程的更新

3. **变更日志** (`CHANGELOG.md`)
   - 功能新增、修改、删除
   - 破坏性变更说明
   - 版本兼容性信息

4. **架构文档** (`docs/ARCHITECTURE.md`)
   - 新增的模块或服务
   - 数据库结构变更
   - 系统架构调整

### 文档更新模板
```markdown
## [功能名称] - [日期]

### 变更类型
- [ ] 新增功能
- [ ] 功能修改
- [ ] 问题修复
- [ ] 性能优化

### 变更描述
简要描述本次变更的内容和目的

### 影响范围
- 前端组件：[列出受影响的组件]
- 后端API：[列出受影响的接口]
- 数据库：[列出数据库变更]

### 使用示例
```typescript
// 提供新功能的使用示例
```

### 注意事项
列出使用时需要注意的事项或限制
```

## 🚫 禁止行为

### 代码层面
- ❌ 复制粘贴现有组件代码创建"新"组件
- ❌ 为了微小差异重新实现整个功能
- ❌ 不查阅文档就开始编码
- ❌ 添加功能相同的重复依赖

### 文档层面
- ❌ 代码变更后不更新文档
- ❌ 只更新代码注释不更新用户文档
- ❌ 文档描述与实际实现不符

## ✅ 推荐行为

### 开发实践
- ✅ 开发前先浏览现有代码库
- ✅ 查阅项目的设计文档和架构说明
- ✅ 与团队成员讨论设计方案
- ✅ 编写可复用的通用组件

### 文档实践
- ✅ 及时更新相关文档
- ✅ 提供清晰的使用示例
- ✅ 记录设计决策和权衡考虑
- ✅ 定期审查和优化文档结构

## 🔍 代码审查要点

### 审查者检查清单
- [ ] 是否复用了现有组件和工具
- [ ] 是否遵循了项目的架构模式
- [ ] 新增依赖是否必要且合理
- [ ] 相关文档是否已更新
- [ ] 代码是否具有良好的可复用性

### 常见问题及解决方案
1. **组件功能重复**
   - 问题：创建了与现有组件功能重复的新组件
   - 解决：合并功能或扩展现有组件

2. **依赖冗余**
   - 问题：添加了功能重复的依赖包
   - 解决：使用现有依赖或替换为更通用的依赖

3. **文档滞后**
   - 问题：代码已更新但文档未同步
   - 解决：建立代码变更与文档更新的关联机制

## 📈 持续改进

### 定期审查
- 每月审查项目依赖，清理不必要的包
- 每季度审查组件库，合并相似功能
- 每半年审查文档结构，优化组织方式

### 团队培训
- 定期分享项目架构和设计模式
- 组织代码复用最佳实践的技术分享
- 建立新人入职时的代码库导览机制

---

**记住：好的代码不仅仅是能工作的代码，更是易于维护、复用和扩展的代码。**
