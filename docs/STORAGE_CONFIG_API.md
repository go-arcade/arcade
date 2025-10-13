# 存储配置 API 使用指南

## 概述

Arcade CI/CD 平台的存储配置已从配置文件迁移到数据库管理，支持动态配置和切换不同的存储后端。

## API 端点

### 1. 创建存储配置

**POST** `/api/v1/storage/configs`

```json
{
  "storageId": "storage_minio_001",
  "name": "MinIO 默认存储",
  "storageType": "minio",
  "config": {
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "bucket": "arcade",
    "region": "us-east-1",
    "useTLS": false,
    "basePath": "artifacts"
  },
  "description": "本地 MinIO 存储配置",
  "isDefault": 1
}
```

### 2. 获取存储配置列表

**GET** `/api/v1/storage/configs`

响应示例：
```json
{
  "code": 200,
  "message": "Storage configs retrieved successfully",
  "data": [
    {
      "storageId": "storage_minio_001",
      "name": "MinIO 默认存储",
      "storageType": "minio",
      "config": {
        "endpoint": "localhost:9000",
        "accessKey": "minioadmin",
        "secretKey": "minioadmin",
        "bucket": "arcade",
        "region": "us-east-1",
        "useTLS": false,
        "basePath": "artifacts"
      },
      "description": "本地 MinIO 存储配置",
      "isDefault": 1,
      "isEnabled": 1
    }
  ]
}
```

### 3. 获取特定存储配置

**GET** `/api/v1/storage/configs/{id}`

### 4. 更新存储配置

**PUT** `/api/v1/storage/configs/{id}`

```json
{
  "name": "MinIO 生产环境存储",
  "storageType": "minio",
  "config": {
    "endpoint": "minio.example.com:9000",
    "accessKey": "production_key",
    "secretKey": "production_secret",
    "bucket": "arcade-prod",
    "region": "us-east-1",
    "useTLS": true,
    "basePath": "artifacts"
  },
  "description": "生产环境 MinIO 存储配置",
  "isDefault": 1,
  "isEnabled": 1
}
```

### 5. 删除存储配置

**DELETE** `/api/v1/storage/configs/{id}`

### 6. 设置默认存储配置

**POST** `/api/v1/storage/configs/{id}/default`

### 7. 获取默认存储配置

**GET** `/api/v1/storage/configs/default`

## 支持的存储类型

### 1. MinIO

```json
{
  "storageType": "minio",
  "config": {
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "bucket": "arcade",
    "region": "us-east-1",
    "useTLS": false,
    "basePath": "artifacts"
  }
}
```

### 2. AWS S3

```json
{
  "storageType": "s3",
  "config": {
    "endpoint": "https://s3.amazonaws.com",
    "accessKey": "AKIAIOSFODNN7EXAMPLE",
    "secretKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
    "bucket": "arcade-artifacts",
    "region": "us-east-1",
    "useTLS": true,
    "basePath": "ci-artifacts"
  }
}
```

### 3. 阿里云 OSS

```json
{
  "storageType": "oss",
  "config": {
    "endpoint": "oss-cn-hangzhou.aliyuncs.com",
    "accessKey": "LTAI4G...",
    "secretKey": "xxx...",
    "bucket": "arcade-artifacts",
    "region": "cn-hangzhou",
    "useTLS": true,
    "basePath": "build-artifacts"
  }
}
```

### 4. Google Cloud Storage

```json
{
  "storageType": "gcs",
  "config": {
    "endpoint": "https://storage.googleapis.com",
    "accessKey": "/path/to/service-account-key.json",
    "bucket": "arcade-artifacts",
    "region": "us-central1",
    "basePath": "ci-builds"
  }
}
```

### 5. 腾讯云 COS

```json
{
  "storageType": "cos",
  "config": {
    "endpoint": "https://cos.ap-guangzhou.myqcloud.com",
    "accessKey": "AKID...",
    "secretKey": "xxx...",
    "bucket": "arcade-artifacts",
    "region": "ap-guangzhou",
    "useTLS": true,
    "basePath": "pipeline-artifacts"
  }
}
```

## 使用示例

### 1. 初始化 MinIO 存储配置

```bash
curl -X POST http://localhost:8080/api/v1/storage/configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "storageId": "storage_minio_001",
    "name": "MinIO 默认存储",
    "storageType": "minio",
    "config": {
      "endpoint": "localhost:9000",
      "accessKey": "minioadmin",
      "secretKey": "minioadmin",
      "bucket": "arcade",
      "region": "us-east-1",
      "useTLS": false,
      "basePath": "artifacts"
    },
    "description": "本地 MinIO 存储配置",
    "isDefault": 1
  }'
```

### 2. 切换到 AWS S3

```bash
# 1. 创建 S3 配置
curl -X POST http://localhost:8080/api/v1/storage/configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "storageId": "storage_s3_001",
    "name": "AWS S3 存储",
    "storageType": "s3",
    "config": {
      "endpoint": "https://s3.amazonaws.com",
      "accessKey": "YOUR_ACCESS_KEY",
      "secretKey": "YOUR_SECRET_KEY",
      "bucket": "arcade-artifacts",
      "region": "us-east-1",
      "useTLS": true,
      "basePath": "ci-artifacts"
    },
    "description": "AWS S3 存储配置",
    "isDefault": 0
  }'

# 2. 设置为默认配置
curl -X POST http://localhost:8080/api/v1/storage/configs/storage_s3_001/default \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 3. 获取当前默认存储配置

```bash
curl -X GET http://localhost:8080/api/v1/storage/configs/default \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 配置验证

系统会自动验证存储配置的有效性：

- **必填字段检查**: endpoint、accessKey、secretKey、bucket
- **存储类型验证**: 确保配置格式符合对应存储类型的要求
- **连接测试**: 创建配置时会测试存储连接是否可用

## 安全注意事项

1. **敏感信息加密**: 存储的密钥信息在数据库中会被加密存储
2. **访问控制**: 存储配置管理需要管理员权限
3. **审计日志**: 所有存储配置的变更都会记录审计日志

## 迁移指南

### 从配置文件迁移

1. **备份现有配置**: 记录当前配置文件中的存储设置
2. **创建数据库配置**: 使用 API 创建对应的存储配置
3. **验证功能**: 确保新的存储配置工作正常
4. **移除配置文件**: 删除配置文件中的 `[storage]` 部分

### 配置对比

**旧配置 (config.toml)**:
```toml
[storage]
Provider  = "minio"
AccessKey = "minioadmin"
SecretKey = "minioadmin"
Endpoint  = "localhost:9000"
Bucket    = "arcade"
Region    = "us-east-1"
UseTLS    = false
BasePath  = "artifacts"
```

**新配置 (数据库)**:
```json
{
  "storageId": "storage_minio_001",
  "name": "MinIO 默认存储",
  "storageType": "minio",
  "config": {
    "endpoint": "localhost:9000",
    "accessKey": "minioadmin",
    "secretKey": "minioadmin",
    "bucket": "arcade",
    "region": "us-east-1",
    "useTLS": false,
    "basePath": "artifacts"
  },
  "isDefault": 1
}
```

## 故障排除

### 常见问题

1. **存储连接失败**: 检查网络连接和认证信息
2. **权限不足**: 确保存储账号有足够的读写权限
3. **配置格式错误**: 验证 JSON 格式和必填字段

### 调试方法

1. **查看存储配置**: 使用 GET API 检查当前配置
2. **测试连接**: 使用存储客户端工具测试连接
3. **查看日志**: 检查应用日志中的存储相关错误

## 最佳实践

1. **多环境配置**: 为不同环境创建不同的存储配置
2. **定期备份**: 定期备份存储配置和重要数据
3. **监控告警**: 设置存储使用量和性能监控
4. **权限最小化**: 使用最小权限原则配置存储访问
