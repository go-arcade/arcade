# Python Notify Plugin

这是一个使用 Python 开发的 Arcade 通知插件示例。

## 功能

- ✅ 发送单条消息
- ✅ 发送模板消息  
- ✅ 批量发送消息
- ✅ 支持自定义配置

## 环境要求

- Python 3.7+
- pip

## 安装依赖

```bash
pip install -r requirements.txt
```

## 本地测试

### 1. 设置环境变量

```bash
export ARCADE_RPC_PLUGIN="arcade-rpc-plugin-protocol"
export PLUGIN_PROTOCOL_VERSIONS="2"
```

### 2. 运行插件

```bash
python3 plugin.py
```

插件启动后会输出握手信息，例如：
```
1|2|tcp|127.0.0.1:54321|grpc
```

## 打包插件

```bash
# 确保 plugin.py 有执行权限
chmod +x plugin.py

# 打包为 zip
zip python-notify-plugin.zip plugin.py manifest.json requirements.txt README.md
```

## 安装到 Arcade

### 使用 API

```bash
curl -X POST http://localhost:8080/api/v1/plugins/install \
  -F "source=local" \
  -F "file=@python-notify-plugin.zip"
```

### 配置插件

```bash
curl -X POST http://localhost:8080/api/v1/plugins/<plugin_id>/configs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "生产环境通知",
    "configItems": {
      "webhook_url": "https://your-webhook-url.com/notify",
      "timeout": 60,
      "retry_count": 3
    },
    "scope": "global",
    "isDefault": 1
  }'
```

## 在任务中使用

```yaml
stages:
  - name: build
    jobs:
      - name: build-app
        commands:
          - go build
        plugins:
          - plugin_id: python-notify
            execution_stage: on_success
            params:
              message: "构建成功！"
```

## 自定义开发

### 修改消息发送逻辑

在 `plugin.py` 的 `Send` 方法中实现您的逻辑：

```python
def Send(self, message_json, opts_json):
    try:
        message = json.loads(message_json) if message_json else ""
        
        # ===== 在这里实现您的逻辑 =====
        # 示例1: 发送到钉钉
        import requests
        webhook_url = self.config.get('webhook_url')
        response = requests.post(
            webhook_url,
            json={
                "msgtype": "text",
                "text": {
                    "content": message
                }
            }
        )
        response.raise_for_status()
        
        # 示例2: 发送到企业微信
        # ...
        
        # 示例3: 发送邮件
        # ...
        
        return None  # 成功
        
    except Exception as e:
        return str(e)  # 失败
```

## 支持的 RPC 方法

### 基础方法（所有插件必须实现）

- `Name()` - 返回插件名称
- `Description()` - 返回插件描述
- `Version()` - 返回插件版本
- `Type()` - 返回插件类型
- `Init(config_json)` - 初始化插件
- `Cleanup()` - 清理资源

### Notify 插件方法

- `Send(message_json, opts_json)` - 发送消息
- `SendTemplate(template, data_json, opts_json)` - 发送模板消息
- `SendBatch(messages_json, opts_json)` - 批量发送消息

## 调试技巧

### 查看日志

插件的 stderr 输出会被主程序捕获：

```python
# 输出调试信息
self._debug("这是调试信息")

# 或直接使用
print("调试信息", file=sys.stderr, flush=True)
```

### 单独运行测试

```python
# 在文件末尾添加测试代码
if __name__ == '__main__':
    # 设置环境变量
    os.environ['ARCADE_RPC_PLUGIN'] = 'arcade-rpc-plugin-protocol'
    os.environ['PLUGIN_PROTOCOL_VERSIONS'] = '2'
    
    # 运行主程序
    main()
```

## 常见问题

### Q: 环境变量检查失败

**A:** 确保插件由 Arcade 主程序启动，或手动设置环境变量：
```bash
export ARCADE_RPC_PLUGIN="arcade-rpc-plugin-protocol"
export PLUGIN_PROTOCOL_VERSIONS="2"
```

### Q: 如何添加第三方库

**A:** 在 `requirements.txt` 中添加依赖，然后安装：
```bash
echo "requests>=2.25.0" >> requirements.txt
pip install -r requirements.txt
```

### Q: 如何处理超时

**A:** 在方法中添加超时处理：
```python
import signal

def timeout_handler(signum, frame):
    raise TimeoutError("操作超时")

signal.signal(signal.SIGALRM, timeout_handler)
signal.alarm(30)  # 30秒超时
try:
    # 执行操作
    pass
finally:
    signal.alarm(0)  # 取消超时
```

## 参考资料

- [多语言插件开发指南](../../docs/PLUGIN_MULTI_LANGUAGE.md)
- [插件开发指南](../../docs/PLUGIN_DEVELOPMENT.md)
- [插件快速入门](../../docs/PLUGIN_QUICKSTART_CN.md)

## 许可证

MIT License

