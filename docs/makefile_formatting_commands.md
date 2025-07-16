# Makefile 代码格式化命令文档

## 概述

为项目 Makefile 新增了全面的代码格式化和质量检查命令，支持代码换行、格式化、import 排序等功能。

## 新增命令

### 1. 基础格式化命令

#### `make fmt`
```bash
make fmt
```
- **功能**: 使用 Go 官方 `go fmt` 工具格式化代码
- **作用**: 统一代码缩进、空格、换行等基础格式
- **适用**: 所有 `.go` 文件

#### `make imports`
```bash
make imports
```
- **功能**: 使用 `goimports` 工具排序和格式化 import 语句
- **作用**: 
  - 自动添加缺失的 import
  - 移除未使用的 import
  - 按照标准库、第三方库、本地包的顺序排序
- **本地包前缀**: `ai-svc`
- **自动安装**: 如果未安装会自动安装 `goimports`

#### `make lines`
```bash
make lines
```
- **功能**: 使用 `golines` 工具控制代码行长度
- **作用**:
  - 将超过 120 字符的长行自动换行
  - 保持代码可读性
  - 智能处理函数调用、结构体等
- **配置**:
  - 最大行长度: 120 字符
  - 基础格式化器: `gofumpt`
- **自动安装**: 如果未安装会自动安装 `golines`

### 2. 代码质量检查命令

#### `make lint`
```bash
make lint
```
- **功能**: 使用 `golangci-lint` 进行代码质量检查
- **检查项目**:
  - 错误处理检查
  - 代码风格检查
  - 安全漏洞检查
  - 性能问题检查
  - 代码复杂度检查
- **配置文件**: `.golangci.yml`
- **自动安装**: 如果未安装会自动安装 `golangci-lint`

#### `make lint-fix`
```bash
make lint-fix
```
- **功能**: 自动修复可修复的 lint 问题
- **作用**: 自动修复简单的代码质量问题
- **注意**: 无法修复的问题需要手动处理

### 3. 组合命令

#### `make format`
```bash
make format
```
- **功能**: 执行完整的代码格式化流程
- **执行顺序**:
  1. `make fmt` - 基础格式化
  2. `make imports` - Import 排序
  3. `make lines` - 行长度控制
  4. `make lint-fix` - 修复 lint 问题
- **推荐使用**: 提交代码前的标准格式化流程

#### `make check`
```bash
make check
```
- **功能**: 完整的代码质量检查
- **执行顺序**:
  1. `make format` - 完整格式化
  2. `make test` - 运行测试
  3. `make lint` - 代码质量检查
- **用途**: CI/CD 流程或代码审查前的全面检查

### 4. 工具管理命令

#### `make install-tools`
```bash
make install-tools
```
- **功能**: 安装所有必需的开发工具
- **安装工具**:
  - `goimports`: Import 管理工具
  - `golines`: 行长度控制工具
  - `golangci-lint`: 代码质量检查工具
- **用途**: 新环境设置或工具更新

## 配置文件

### `.golangci.yml`
完整的 golangci-lint 配置文件，包含：

#### 启用的检查器
```yaml
linters:
  enable:
    - errcheck      # 错误处理检查
    - gosimple      # 代码简化
    - govet         # Go 官方检查
    - staticcheck   # 静态分析
    - gocyclo       # 复杂度检查
    - gofmt         # 格式检查
    - goimports     # Import 检查
    - gosec         # 安全检查
    - lll           # 行长度检查
    # ... 更多检查器
```

#### 行长度设置
```yaml
linters-settings:
  lll:
    line-length: 120
    tab-width: 4
```

#### 排除规则
- 测试文件排除某些检查
- 模型文件排除行长度检查（GORM 标签）
- 中间件文件排除行长度检查（CORS 头部）
- CMD 层排除错误检查

## 使用示例

### 日常开发工作流

1. **开发过程中**:
   ```bash
   # 保存文件后格式化
   make fmt
   
   # 添加新依赖后
   make imports
   
   # 代码行太长时
   make lines
   ```

2. **提交前**:
   ```bash
   # 完整格式化
   make format
   
   # 运行测试
   make test
   ```

3. **代码审查前**:
   ```bash
   # 全面检查
   make check
   ```

### CI/CD 集成

```yaml
# GitHub Actions 示例
- name: 代码质量检查
  run: |
    make install-tools
    make check
```

### IDE 集成

#### VS Code
```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "go.lintOnSave": "package"
}
```

## 工具特性

### golines 特性
- **智能换行**: 在合适的位置换行，不破坏代码逻辑
- **保持可读性**: 优先保证代码可读性
- **配置灵活**: 支持自定义行长度和基础格式化器

### goimports 特性
- **自动 import**: 自动添加缺失的依赖
- **智能排序**: 按照 Go 社区标准排序
- **本地包识别**: 正确处理项目内部包

### golangci-lint 特性
- **多检查器**: 集成多个静态分析工具
- **可配置**: 支持细粒度配置
- **性能优化**: 并行执行多个检查器

## 最佳实践

### 1. 提交前检查
```bash
# 标准流程
make format   # 格式化代码
make test     # 运行测试
git add .     # 添加更改
git commit    # 提交
```

### 2. 团队规范
- 所有团队成员使用相同的格式化配置
- 在 pre-commit hook 中集成格式化检查
- CI/CD 中强制执行代码质量检查

### 3. 性能考虑
- `make format` 可能耗时较长，建议在提交前执行
- 日常开发可以使用单独的命令（`make fmt`）
- 大型项目可以考虑增量检查

## 故障排除

### 常见问题

1. **工具安装失败**
   ```bash
   # 手动安装
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/segmentio/golines@latest
   ```

2. **权限问题**
   ```bash
   # 确保 GOPATH/bin 在 PATH 中
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

3. **配置冲突**
   - 检查 `.golangci.yml` 配置
   - 确认 IDE 设置不冲突

### 调试命令
```bash
# 查看工具版本
goimports -version
golines --version
golangci-lint --version

# 测试单个文件
goimports -d file.go
golines -m 120 file.go
golangci-lint run file.go
```

## 总结

新增的格式化命令提供了：

1. **完整的代码格式化工具链**
2. **灵活的使用方式**（单独命令 + 组合命令）
3. **自动工具安装**
4. **详细的配置选项**
5. **团队协作支持**

通过这些工具，可以确保代码库的一致性和质量，提高团队开发效率。 