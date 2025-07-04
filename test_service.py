#!/usr/bin/env python3
"""
AI API Gateway 服务测试脚本
测试各种API端点的功能和响应
"""

import requests
import json
import time
import sys
from typing import Dict, Any, Optional

class APIGatewayTester:
    def __init__(self, base_url: str = "http://localhost:8080", api_key: str = None):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        
        # 设置默认headers
        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}',
                'Content-Type': 'application/json'
            })
    
    def print_test_header(self, test_name: str):
        """打印测试标题"""
        print(f"\n{'='*60}")
        print(f"🧪 测试: {test_name}")
        print(f"{'='*60}")
    
    def print_result(self, success: bool, message: str, details: str = None):
        """打印测试结果"""
        status = "✅ 成功" if success else "❌ 失败"
        print(f"{status}: {message}")
        if details:
            print(f"   详情: {details}")
    
    def make_request(self, method: str, endpoint: str, data: Dict = None, headers: Dict = None) -> tuple:
        """发送HTTP请求"""
        url = f"{self.base_url}{endpoint}"
        
        try:
            if method.upper() == 'GET':
                response = self.session.get(url, headers=headers)
            elif method.upper() == 'POST':
                response = self.session.post(url, json=data, headers=headers)
            elif method.upper() == 'PUT':
                response = self.session.put(url, json=data, headers=headers)
            elif method.upper() == 'DELETE':
                response = self.session.delete(url, headers=headers)
            else:
                return False, f"不支持的HTTP方法: {method}"
            
            return True, response
        except requests.exceptions.ConnectionError:
            return False, "连接失败 - 服务器可能未启动"
        except requests.exceptions.Timeout:
            return False, "请求超时"
        except Exception as e:
            return False, f"请求异常: {str(e)}"
    
    def test_health_check(self):
        """测试健康检查"""
        self.print_test_header("健康检查")
        
        success, result = self.make_request('GET', '/health')
        if not success:
            self.print_result(False, "健康检查失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            try:
                data = response.json()
                status = data.get('data', {}).get('status', 'unknown')
                self.print_result(True, f"健康检查通过 (状态: {status})")
                print(f"   响应时间: {response.elapsed.total_seconds():.3f}s")
                return True
            except json.JSONDecodeError:
                self.print_result(False, "响应格式错误", "无法解析JSON")
                return False
        else:
            self.print_result(False, f"健康检查失败 (状态码: {response.status_code})")
            return False
    
    def test_readiness_check(self):
        """测试就绪检查"""
        self.print_test_header("就绪检查")
        
        success, result = self.make_request('GET', '/health/ready')
        if not success:
            self.print_result(False, "就绪检查失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            self.print_result(True, "就绪检查通过")
            return True
        else:
            self.print_result(False, f"就绪检查失败 (状态码: {response.status_code})")
            return False
    
    def test_metrics(self):
        """测试监控指标"""
        self.print_test_header("监控指标")
        
        success, result = self.make_request('GET', '/metrics')
        if not success:
            self.print_result(False, "获取监控指标失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            metrics_count = len(response.text.split('\n'))
            self.print_result(True, f"监控指标获取成功 ({metrics_count} 行)")
            return True
        else:
            self.print_result(False, f"获取监控指标失败 (状态码: {response.status_code})")
            return False
    
    def test_models_api(self):
        """测试模型列表API"""
        self.print_test_header("模型列表API")
        
        if not self.api_key:
            self.print_result(False, "需要API密钥", "请提供有效的API密钥")
            return False
        
        success, result = self.make_request('GET', '/v1/models')
        if not success:
            self.print_result(False, "获取模型列表失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            try:
                data = response.json()
                models = data.get('data', [])
                self.print_result(True, f"获取模型列表成功 ({len(models)} 个模型)")
                for model in models[:3]:  # 显示前3个模型
                    print(f"   - {model.get('id', 'unknown')}")
                return True
            except json.JSONDecodeError:
                self.print_result(False, "响应格式错误", "无法解析JSON")
                return False
        elif response.status_code == 401:
            self.print_result(False, "认证失败", "API密钥无效或已过期")
            return False
        else:
            self.print_result(False, f"获取模型列表失败 (状态码: {response.status_code})")
            return False
    
    def test_chat_completions(self):
        """测试聊天完成API"""
        self.print_test_header("聊天完成API")
        
        if not self.api_key:
            self.print_result(False, "需要API密钥", "请提供有效的API密钥")
            return False
        
        test_data = {
            "model": "gpt-3.5-turbo",
            "messages": [
                {"role": "user", "content": "Hello! This is a test message."}
            ],
            "max_tokens": 50
        }
        
        success, result = self.make_request('POST', '/v1/chat/completions', test_data)
        if not success:
            self.print_result(False, "聊天完成请求失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            try:
                data = response.json()
                choices = data.get('choices', [])
                if choices:
                    message = choices[0].get('message', {}).get('content', '')
                    self.print_result(True, "聊天完成成功")
                    print(f"   响应: {message[:100]}...")
                else:
                    self.print_result(False, "响应格式异常", "没有找到choices")
                return True
            except json.JSONDecodeError:
                self.print_result(False, "响应格式错误", "无法解析JSON")
                return False
        elif response.status_code == 401:
            self.print_result(False, "认证失败", "API密钥无效或已过期")
            return False
        elif response.status_code == 400:
            self.print_result(False, "请求参数错误", "检查请求格式")
            return False
        else:
            self.print_result(False, f"聊天完成失败 (状态码: {response.status_code})")
            try:
                error_data = response.json()
                print(f"   错误信息: {error_data}")
            except:
                print(f"   响应内容: {response.text[:200]}")
            return False
    
    def test_admin_apis(self):
        """测试管理API"""
        self.print_test_header("管理API测试")
        
        # 测试获取用户列表
        success, result = self.make_request('GET', '/admin/users/')
        if not success:
            self.print_result(False, "获取用户列表失败", result)
            return False
        
        response = result
        if response.status_code == 200:
            try:
                data = response.json()
                users = data.get('data', {}).get('items', [])
                self.print_result(True, f"获取用户列表成功 ({len(users)} 个用户)")
                return True
            except json.JSONDecodeError:
                self.print_result(False, "响应格式错误", "无法解析JSON")
                return False
        else:
            self.print_result(False, f"获取用户列表失败 (状态码: {response.status_code})")
            return False
    
    def run_all_tests(self):
        """运行所有测试"""
        print("🚀 开始测试AI API Gateway服务")
        print(f"📍 服务地址: {self.base_url}")
        if self.api_key:
            print(f"🔑 API密钥: {self.api_key[:10]}...")
        else:
            print("⚠️  未提供API密钥，部分测试将跳过")
        
        tests = [
            ("基础连通性", self.test_health_check),
            ("就绪状态", self.test_readiness_check),
            ("监控指标", self.test_metrics),
            ("模型列表", self.test_models_api),
            ("聊天完成", self.test_chat_completions),
            ("管理接口", self.test_admin_apis),
        ]
        
        results = []
        for test_name, test_func in tests:
            try:
                result = test_func()
                results.append((test_name, result))
            except Exception as e:
                self.print_result(False, f"{test_name}测试异常", str(e))
                results.append((test_name, False))
        
        # 打印总结
        self.print_test_header("测试总结")
        passed = sum(1 for _, result in results if result)
        total = len(results)
        
        print(f"📊 测试结果: {passed}/{total} 通过")
        
        for test_name, result in results:
            status = "✅" if result else "❌"
            print(f"   {status} {test_name}")
        
        if passed == total:
            print("\n🎉 所有测试通过！服务运行正常。")
            return True
        else:
            print(f"\n⚠️  有 {total - passed} 个测试失败，请检查服务配置。")
            return False

def main():
    """主函数"""
    # 默认配置
    base_url = "http://localhost:8080"
    api_key = "ak_4278fc65b1d32cc99fe69fc25bf352261fab3aa0b08488d919dce0097b0f3915"
    
    # 可以通过命令行参数覆盖
    if len(sys.argv) > 1:
        base_url = sys.argv[1]
    if len(sys.argv) > 2:
        api_key = sys.argv[2]
    
    # 创建测试器并运行测试
    tester = APIGatewayTester(base_url, api_key)
    success = tester.run_all_tests()
    
    # 退出码
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()
