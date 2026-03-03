# GoSQL-Porter

基于 MySQL Shell 的数据库迁移工具，支持数据库导出、导入和备份功能。

## 功能特性

- 支持多数据库批量导出/导入
- 支持多线程并行处理
- 支持导出文件自动备份（zip 格式）
- 支持 Linux AMD64/ARM64、Windows AMD64 平台

## 环境要求

- Go 1.25+
- MySQL Shell 8.0+（需添加到系统环境变量）

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone <repository-url>
cd gosql-porter

# 编译
go build -o gosql-porter ./main.go
```

### 跨平台编译

运行打包脚本：

```bash
# Windows
build.bat

# 输出目录: build/
# - gosql-porter-linux-amd64
# - gosql-porter-linux-arm64
# - gosql-porter-windows-amd64.exe
```

## 使用方法

```bash
# 使用默认配置文件 (config.yml)
gosql-porter

# 指定配置文件
gosql-porter -c /path/to/config.yml
```

## 配置文件

创建 `config.yml` 配置文件：

```yaml
db_settings:
  source:
    host: "10.6.0.100"
    port: 3306
    username: "root"
    passwd: "password"
  target:
    host: "10.6.0.200"
    port: 3306
    username: "root"
    passwd: "password"

options:
  # 要迁移的数据库名称列表
  databases:
    - "db1"
    - "db2"
    - "db3"

  # SQL 导出目录（默认: dumpSql）
  dump_to: "./dumpSql"

  # 备份文件保存目录（可选，不配置则不备份）
  save_to: "./backup"

  # 并发线程数（默认: 4）
  threads: 16

  # 运行模式
  # 0: 导出 + 导入（默认）
  # 1: 仅导出
  # 2: 仅导入
  mode: 0
```

### 配置项说明

| 配置项               | 类型     | 必填 | 说明                     |
| -------------------- | -------- | ---- | ------------------------ |
| `db_settings.source` | object   | 是   | 源数据库连接配置         |
| `db_settings.target` | object   | 是   | 目标数据库连接配置       |
| `options.databases`  | []string | 是   | 要迁移的数据库列表       |
| `options.dump_to`    | string   | 否   | 导出目录，默认 `dumpSql` |
| `options.save_to`    | string   | 否   | 备份目录，不配置则不备份 |
| `options.threads`    | int      | 否   | 并发线程数，默认 4       |
| `options.mode`       | int      | 否   | 运行模式，默认 0         |

### 运行模式

| 模式值 | 说明                             |
| ------ | -------------------------------- |
| `0`    | 导出 + 导入（完整迁移流程）      |
| `1`    | 仅导出（从源数据库导出到文件）   |
| `2`    | 仅导入（从文件导入到目标数据库） |

## 工作流程

1. **连接检查** - 验证 MySQL Shell 连接是否可用
2. **导出** - 使用 `util.dumpSchemas()` 导出指定数据库
3. **备份** - 将导出文件打包为 zip（可选）
4. **导入** - 删除目标数据库后使用 `util.loadDump()` 导入

## 注意事项

- 导入前会**删除目标数据库**中已存在的同名数据库
- 密码支持特殊字符（自动 URL 编码）
- 确保 MySQL Shell 已安装并在 PATH 中可用

## 依赖

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- [github.com/spf13/viper](https://github.com/spf13/viper) - 配置管理

## License

MIT
