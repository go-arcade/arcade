#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
测试插件脚本
用于在不启动 Arcade 主程序的情况下测试插件
"""

import json
import sys
import os

# 添加当前目录到路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from plugin import NotifyPlugin


def test_plugin_basic():
    """测试插件基本功能"""
    print("=== 测试插件基本功能 ===\n")
    
    # 创建插件实例
    plugin = NotifyPlugin()
    
    # 测试 Name
    name = plugin.Name()
    print(f"✓ Name(): {name}")
    assert name == "python-notify", f"Expected 'python-notify', got '{name}'"
    
    # 测试 Description
    desc = plugin.Description()
    print(f"✓ Description(): {desc}")
    
    # 测试 Version
    version = plugin.Version()
    print(f"✓ Version(): {version}")
    assert version == "1.0.0", f"Expected '1.0.0', got '{version}'"
    
    # 测试 Type
    plugin_type = plugin.Type()
    print(f"✓ Type(): {plugin_type}")
    assert plugin_type == "notify", f"Expected 'notify', got '{plugin_type}'"
    
    print("\n✅ 所有基本测试通过！\n")


def test_plugin_init():
    """测试插件初始化"""
    print("=== 测试插件初始化 ===\n")
    
    plugin = NotifyPlugin()
    
    # 测试空配置
    error = plugin.Init(None)
    assert error is None, f"Init with None config failed: {error}"
    print("✓ Init with None config")
    
    # 测试有效配置
    config = {
        "webhook_url": "https://example.com/webhook",
        "timeout": 60,
        "retry_count": 3
    }
    config_json = json.dumps(config)
    
    error = plugin.Init(config_json)
    assert error is None, f"Init with valid config failed: {error}"
    print(f"✓ Init with config: {config}")
    
    # 验证配置已保存
    assert plugin.config == config, "Config not saved correctly"
    print("✓ Config saved correctly")
    
    # 测试无效配置
    plugin2 = NotifyPlugin()
    error = plugin2.Init("{invalid json")
    assert error is not None, "Init should fail with invalid JSON"
    print(f"✓ Init correctly handles invalid JSON: {error}")
    
    print("\n✅ 所有初始化测试通过！\n")


def test_plugin_send():
    """测试发送消息"""
    print("=== 测试发送消息 ===\n")
    
    plugin = NotifyPlugin()
    
    # 初始化配置
    config = {"webhook_url": "https://example.com/webhook"}
    plugin.Init(json.dumps(config))
    
    # 测试发送消息
    message = "测试消息"
    message_json = json.dumps(message)
    
    error = plugin.Send(message_json, None)
    assert error is None, f"Send failed: {error}"
    print(f"✓ Send message: {message}")
    
    # 测试空消息
    error = plugin.Send(None, None)
    assert error is None, f"Send with None message failed: {error}"
    print("✓ Send with None message")
    
    print("\n✅ 所有发送测试通过！\n")


def test_plugin_send_template():
    """测试发送模板消息"""
    print("=== 测试发送模板消息 ===\n")
    
    plugin = NotifyPlugin()
    plugin.Init(None)
    
    # 测试模板消息
    template = "Hello, {{name}}!"
    data = {"name": "World"}
    data_json = json.dumps(data)
    
    error = plugin.SendTemplate(template, data_json, None)
    assert error is None, f"SendTemplate failed: {error}"
    print(f"✓ SendTemplate: template={template}, data={data}")
    
    print("\n✅ 所有模板测试通过！\n")


def test_plugin_send_batch():
    """测试批量发送"""
    print("=== 测试批量发送 ===\n")
    
    plugin = NotifyPlugin()
    plugin.Init(None)
    
    # 测试批量消息
    messages = ["消息1", "消息2", "消息3"]
    messages_json = json.dumps(messages)
    
    error = plugin.SendBatch(messages_json, None)
    assert error is None, f"SendBatch failed: {error}"
    print(f"✓ SendBatch: {len(messages)} messages")
    
    print("\n✅ 所有批量测试通过！\n")


def test_plugin_cleanup():
    """测试清理"""
    print("=== 测试清理 ===\n")
    
    plugin = NotifyPlugin()
    plugin.Init(None)
    
    error = plugin.Cleanup()
    assert error is None, f"Cleanup failed: {error}"
    print("✓ Cleanup successful")
    
    print("\n✅ 清理测试通过！\n")


def run_all_tests():
    """运行所有测试"""
    print("\n" + "="*60)
    print(" Python 插件单元测试")
    print("="*60 + "\n")
    
    try:
        test_plugin_basic()
        test_plugin_init()
        test_plugin_send()
        test_plugin_send_template()
        test_plugin_send_batch()
        test_plugin_cleanup()
        
        print("\n" + "="*60)
        print(" 🎉 所有测试通过！")
        print("="*60 + "\n")
        
        return 0
        
    except AssertionError as e:
        print(f"\n❌ 测试失败: {e}\n")
        return 1
    except Exception as e:
        print(f"\n❌ 测试错误: {e}\n")
        import traceback
        traceback.print_exc()
        return 1


if __name__ == '__main__':
    sys.exit(run_all_tests())

