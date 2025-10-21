#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Arcade Python Notify Plugin
一个使用 Python 开发的 Arcade 通知插件示例
"""

import sys
import os
import json
import socket
import struct
import threading


class NotifyPlugin:
    """通知插件实现"""
    
    def __init__(self):
        self.name = "python-notify"
        self.version = "1.0.0"
        self.plugin_type = "notify"
        self.description = "Python 通知插件示例"
        self.config = {}
        self._debug("插件实例已创建")
    
    def _debug(self, message):
        """输出调试信息到 stderr"""
        print(f"[{self.name}] {message}", file=sys.stderr, flush=True)
    
    # ========== 基础接口实现 ==========
    
    def Name(self):
        """返回插件名称"""
        self._debug(f"Name() 被调用")
        return self.name
    
    def Description(self):
        """返回插件描述"""
        self._debug(f"Description() 被调用")
        return self.description
    
    def Version(self):
        """返回插件版本"""
        self._debug(f"Version() 被调用")
        return self.version
    
    def Type(self):
        """返回插件类型"""
        self._debug(f"Type() 被调用")
        return self.plugin_type
    
    def Init(self, config_json):
        """
        初始化插件
        
        Args:
            config_json: JSON 格式的配置字符串
        
        Returns:
            None 表示成功，字符串表示错误信息
        """
        try:
            self._debug(f"Init() 被调用，配置: {config_json}")
            
            if config_json:
                # 解码 JSON（如果是 bytes）
                if isinstance(config_json, bytes):
                    config_json = config_json.decode('utf-8')
                
                self.config = json.loads(config_json)
                self._debug(f"配置解析成功: {self.config}")
            
            # 验证配置（示例）
            # if 'webhook_url' not in self.config:
            #     return "webhook_url is required in config"
            
            self._debug("插件初始化成功")
            return None
            
        except json.JSONDecodeError as e:
            error_msg = f"配置解析失败: {str(e)}"
            self._debug(error_msg)
            return error_msg
        except Exception as e:
            error_msg = f"初始化失败: {str(e)}"
            self._debug(error_msg)
            return error_msg
    
    def Cleanup(self):
        """
        清理资源
        
        Returns:
            None 表示成功，字符串表示错误信息
        """
        try:
            self._debug("Cleanup() 被调用")
            
            # 在这里清理资源
            # 例如：关闭连接、释放文件句柄等
            
            self._debug("插件清理成功")
            return None
            
        except Exception as e:
            error_msg = f"清理失败: {str(e)}"
            self._debug(error_msg)
            return error_msg
    
    # ========== Notify 插件接口实现 ==========
    
    def Send(self, message_json, opts_json):
        """
        发送消息
        
        Args:
            message_json: JSON 格式的消息内容
            opts_json: JSON 格式的选项
        
        Returns:
            None 表示成功，字符串表示错误信息
        """
        try:
            self._debug(f"Send() 被调用")
            
            # 解析消息
            if isinstance(message_json, bytes):
                message_json = message_json.decode('utf-8')
            
            message = json.loads(message_json) if message_json else ""
            self._debug(f"发送消息: {message}")
            
            # ===== 在这里实现实际的消息发送逻辑 =====
            # 示例：发送到 Webhook
            # import requests
            # webhook_url = self.config.get('webhook_url')
            # response = requests.post(webhook_url, json={'text': message})
            # response.raise_for_status()
            
            self._debug("消息发送成功")
            return None
            
        except Exception as e:
            error_msg = f"发送消息失败: {str(e)}"
            self._debug(error_msg)
            return error_msg
    
    def SendTemplate(self, template, data_json, opts_json):
        """
        发送模板消息
        
        Args:
            template: 模板字符串
            data_json: JSON 格式的模板数据
            opts_json: JSON 格式的选项
        
        Returns:
            None 表示成功，字符串表示错误信息
        """
        try:
            self._debug(f"SendTemplate() 被调用")
            
            # 解析数据
            if isinstance(data_json, bytes):
                data_json = data_json.decode('utf-8')
            
            data = json.loads(data_json) if data_json else {}
            self._debug(f"发送模板消息: 模板={template}, 数据={data}")
            
            # ===== 在这里实现模板消息发送逻辑 =====
            # 示例：渲染模板并发送
            # from jinja2 import Template
            # t = Template(template)
            # rendered = t.render(data)
            # self.Send(json.dumps(rendered), opts_json)
            
            self._debug("模板消息发送成功")
            return None
            
        except Exception as e:
            error_msg = f"发送模板消息失败: {str(e)}"
            self._debug(error_msg)
            return error_msg
    
    def SendBatch(self, messages_json, opts_json):
        """
        批量发送消息
        
        Args:
            messages_json: JSON 格式的消息列表
            opts_json: JSON 格式的选项
        
        Returns:
            None 表示成功，字符串表示错误信息
        """
        try:
            self._debug(f"SendBatch() 被调用")
            
            # 解析消息列表
            if isinstance(messages_json, bytes):
                messages_json = messages_json.decode('utf-8')
            
            messages = json.loads(messages_json) if messages_json else []
            self._debug(f"批量发送 {len(messages)} 条消息")
            
            # ===== 在这里实现批量发送逻辑 =====
            # 示例：遍历发送
            # for msg in messages:
            #     self.Send(json.dumps(msg), opts_json)
            
            self._debug("批量消息发送成功")
            return None
            
        except Exception as e:
            error_msg = f"批量发送消息失败: {str(e)}"
            self._debug(error_msg)
            return error_msg


class SimpleRPCServer:
    """
    简化的 RPC 服务器
    使用 net/rpc 协议的简化实现
    """
    
    def __init__(self, plugin):
        self.plugin = plugin
        self.server_socket = None
        self.client_socket = None
        self.running = False
    
    def _debug(self, message):
        print(f"[RPC] {message}", file=sys.stderr, flush=True)
    
    def start(self):
        """启动 RPC 服务器"""
        try:
            # 创建 TCP 服务器，监听随机端口
            self.server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            self.server_socket.bind(('127.0.0.1', 0))
            self.server_socket.listen(1)
            
            # 获取实际端口
            host, port = self.server_socket.getsockname()
            
            # 输出握手信息到 stdout
            # 格式: 1|2|tcp|host:port|grpc
            handshake = f"1|2|tcp|{host}:{port}|grpc\n"
            sys.stdout.write(handshake)
            sys.stdout.flush()
            
            self._debug(f"服务器启动在 {host}:{port}")
            self._debug("等待客户端连接...")
            
            # 接受连接
            self.client_socket, addr = self.server_socket.accept()
            self._debug(f"客户端已连接: {addr}")
            
            self.running = True
            
            # 处理请求
            self._handle_requests()
            
        except Exception as e:
            self._debug(f"服务器错误: {e}")
            raise
        finally:
            self.stop()
    
    def _handle_requests(self):
        """处理 RPC 请求"""
        while self.running:
            try:
                # 读取请求
                # net/rpc 格式: [header][body]
                # header: 12 bytes (seq: uint64, service_method_len: uint32)
                
                data = self.client_socket.recv(4096)
                if not data:
                    self._debug("连接关闭")
                    break
                
                # 简化处理：直接读取 JSON-RPC 格式请求
                try:
                    request = json.loads(data.decode('utf-8'))
                except json.JSONDecodeError:
                    self._debug(f"无法解析请求: {data[:100]}")
                    continue
                
                # 提取方法和参数
                method = request.get('method', '')
                params = request.get('params', [])
                req_id = request.get('id', 0)
                
                self._debug(f"收到请求: method={method}, params={params}")
                
                # 调用插件方法
                result, error = self._call_method(method, params)
                
                # 构建响应
                response = {
                    'jsonrpc': '2.0',
                    'id': req_id,
                    'result': result,
                    'error': error
                }
                
                # 发送响应
                response_data = json.dumps(response).encode('utf-8')
                self.client_socket.sendall(response_data)
                
                self._debug(f"响应已发送")
                
            except ConnectionResetError:
                self._debug("连接被重置")
                break
            except Exception as e:
                self._debug(f"处理请求错误: {e}")
                import traceback
                traceback.print_exc(file=sys.stderr)
                break
    
    def _call_method(self, method, params):
        """
        调用插件方法
        
        Returns:
            (result, error) 元组
        """
        try:
            # 移除可能的前缀
            method_name = method.split('.')[-1]
            
            # 检查方法是否存在
            if not hasattr(self.plugin, method_name):
                return None, f"方法 {method_name} 不存在"
            
            # 调用方法
            func = getattr(self.plugin, method_name)
            result = func(*params)
            
            # 处理返回值
            # 如果返回 None，表示成功且无返回值
            # 如果返回字符串，表示错误信息
            # 如果返回其他值，表示成功结果
            if result is None or (isinstance(result, str) and result.startswith("错误") or result.startswith("失败")):
                # 错误情况
                return None, result
            else:
                # 成功情况
                return result, None
            
        except Exception as e:
            import traceback
            error_msg = f"调用方法失败: {str(e)}\n{traceback.format_exc()}"
            self._debug(error_msg)
            return None, error_msg
    
    def stop(self):
        """停止服务器"""
        self.running = False
        
        if self.client_socket:
            try:
                self.client_socket.close()
            except:
                pass
        
        if self.server_socket:
            try:
                self.server_socket.close()
            except:
                pass
        
        self._debug("服务器已停止")


def check_environment():
    """检查环境变量"""
    magic_cookie = os.environ.get('ARCADE_RPC_PLUGIN')
    
    if magic_cookie != 'arcade-rpc-plugin-protocol':
        print("错误: 环境变量 ARCADE_RPC_PLUGIN 不正确", file=sys.stderr)
        print("提示: 此插件必须由 Arcade 主程序启动", file=sys.stderr)
        sys.exit(1)
    
    protocol_versions = os.environ.get('PLUGIN_PROTOCOL_VERSIONS', '1')
    if '2' not in protocol_versions:
        print("错误: 不支持的协议版本", file=sys.stderr)
        sys.exit(1)


def main():
    """主函数"""
    print("[Main] Python 插件启动中...", file=sys.stderr, flush=True)
    
    # 检查环境
    check_environment()
    
    # 创建插件实例
    plugin = NotifyPlugin()
    
    # 创建并启动 RPC 服务器
    server = SimpleRPCServer(plugin)
    
    try:
        server.start()
    except KeyboardInterrupt:
        print("[Main] 收到中断信号", file=sys.stderr)
    except Exception as e:
        print(f"[Main] 错误: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc(file=sys.stderr)
        sys.exit(1)
    finally:
        # 清理插件
        plugin.Cleanup()


if __name__ == '__main__':
    main()

