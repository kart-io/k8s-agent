#!/usr/bin/env python3
"""
K8s Agent Python 客户端 - 查询参数风格

使用示例:
    python python_client.py
"""

import requests
from typing import Dict, Any, Optional
from urllib.parse import urlencode


class K8sClient:
    """K8s Agent 客户端 (查询参数风格)"""

    def __init__(self, base_url: str = "http://localhost:8080"):
        """
        初始化客户端

        Args:
            base_url: API 基础 URL
        """
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
        })

    def _build_url(self, path: str, params: Dict[str, Any]) -> str:
        """
        构建带查询参数的 URL

        Args:
            path: API 路径
            params: 查询参数字典

        Returns:
            完整的 URL
        """
        # 移除空值参数
        clean_params = {k: v for k, v in params.items() if v is not None and v != ''}

        url = f"{self.base_url}{path}"
        if clean_params:
            url += f"?{urlencode(clean_params)}"

        return url

    def _request(self, method: str, path: str, params: Dict[str, Any] = None,
                 json: Dict[str, Any] = None) -> Dict[str, Any]:
        """
        执行 HTTP 请求

        Args:
            method: HTTP 方法
            path: API 路径
            params: 查询参数
            json: JSON 请求体

        Returns:
            API 响应

        Raises:
            requests.HTTPError: HTTP 错误
        """
        params = params or {}
        url = self._build_url(path, params)

        response = self.session.request(method, url, json=json)
        response.raise_for_status()

        return response.json()

    # ===========================
    # 集群管理 API
    # ===========================

    def list_clusters(self, page: int = 1, page_size: int = 20) -> Dict[str, Any]:
        """
        列出所有集群

        Args:
            page: 页码
            page_size: 每页大小

        Returns:
            集群列表响应
        """
        params = {
            'page': page,
            'pageSize': page_size
        }
        return self._request('GET', '/api/k8s/clusters', params)

    def get_cluster(self, cluster_id: str) -> Dict[str, Any]:
        """
        获取集群详情 (查询参数)

        Args:
            cluster_id: 集群 ID

        Returns:
            集群详情
        """
        params = {'clusterId': cluster_id}
        return self._request('GET', '/api/k8s/cluster', params)

    def get_cluster_health(self, cluster_id: str) -> Dict[str, Any]:
        """
        获取集群健康状态 (查询参数)

        Args:
            cluster_id: 集群 ID

        Returns:
            健康状态
        """
        params = {'clusterId': cluster_id}
        return self._request('GET', '/api/k8s/cluster/health', params)

    # ===========================
    # 命名空间管理 API
    # ===========================

    def list_namespaces(self, cluster_id: str) -> Dict[str, Any]:
        """
        列出命名空间 (查询参数)

        Args:
            cluster_id: 集群 ID

        Returns:
            命名空间列表
        """
        params = {'clusterId': cluster_id}
        return self._request('GET', '/api/k8s/namespaces', params)

    def get_namespace(self, cluster_id: str, namespace: str) -> Dict[str, Any]:
        """
        获取命名空间详情 (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间名称

        Returns:
            命名空间详情
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace
        }
        return self._request('GET', '/api/k8s/namespace', params)

    # ===========================
    # Pod 管理 API
    # ===========================

    def list_pods(self, cluster_id: str, namespace: str,
                  page: int = 1, page_size: int = 20) -> Dict[str, Any]:
        """
        列出 Pods (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间
            page: 页码
            page_size: 每页大小

        Returns:
            Pod 列表
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace,
            'page': page,
            'pageSize': page_size
        }
        return self._request('GET', '/api/k8s/pods', params)

    def get_pod(self, cluster_id: str, namespace: str, pod_name: str) -> Dict[str, Any]:
        """
        获取 Pod 详情 (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间
            pod_name: Pod 名称

        Returns:
            Pod 详情
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace,
            'name': pod_name
        }
        return self._request('GET', '/api/k8s/pod', params)

    def get_pod_logs(self, cluster_id: str, namespace: str, pod_name: str,
                     container: Optional[str] = None,
                     tail_lines: Optional[int] = None,
                     follow: bool = False) -> Dict[str, Any]:
        """
        获取 Pod 日志 (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间
            pod_name: Pod 名称
            container: 容器名称 (可选)
            tail_lines: 尾行数 (可选)
            follow: 是否跟踪日志

        Returns:
            Pod 日志
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace,
            'name': pod_name
        }

        if container:
            params['container'] = container
        if tail_lines:
            params['tailLines'] = tail_lines
        if follow:
            params['follow'] = 'true'

        return self._request('GET', '/api/k8s/pod/logs', params)

    # ===========================
    # Deployment 管理 API
    # ===========================

    def list_deployments(self, cluster_id: str, namespace: str) -> Dict[str, Any]:
        """
        列出 Deployments (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间

        Returns:
            Deployment 列表
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace
        }
        return self._request('GET', '/api/k8s/deployments', params)

    def get_deployment(self, cluster_id: str, namespace: str,
                       deployment_name: str) -> Dict[str, Any]:
        """
        获取 Deployment 详情 (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间
            deployment_name: Deployment 名称

        Returns:
            Deployment 详情
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace,
            'name': deployment_name
        }
        return self._request('GET', '/api/k8s/deployment', params)

    # ===========================
    # Node 管理 API
    # ===========================

    def list_nodes(self, cluster_id: str) -> Dict[str, Any]:
        """
        列出 Nodes (查询参数)

        Args:
            cluster_id: 集群 ID

        Returns:
            Node 列表
        """
        params = {'clusterId': cluster_id}
        return self._request('GET', '/api/k8s/nodes', params)

    def get_node(self, cluster_id: str, node_name: str) -> Dict[str, Any]:
        """
        获取 Node 详情 (查询参数)

        Args:
            cluster_id: 集群 ID
            node_name: Node 名称

        Returns:
            Node 详情
        """
        params = {
            'clusterId': cluster_id,
            'name': node_name
        }
        return self._request('GET', '/api/k8s/node', params)

    # ===========================
    # Service 管理 API
    # ===========================

    def list_services(self, cluster_id: str, namespace: str) -> Dict[str, Any]:
        """
        列出 Services (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间

        Returns:
            Service 列表
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace
        }
        return self._request('GET', '/api/k8s/services', params)

    def get_service(self, cluster_id: str, namespace: str,
                    service_name: str) -> Dict[str, Any]:
        """
        获取 Service 详情 (查询参数)

        Args:
            cluster_id: 集群 ID
            namespace: 命名空间
            service_name: Service 名称

        Returns:
            Service 详情
        """
        params = {
            'clusterId': cluster_id,
            'namespace': namespace,
            'name': service_name
        }
        return self._request('GET', '/api/k8s/service', params)


def main():
    """使用示例"""
    # 创建客户端
    client = K8sClient("http://localhost:8080")

    try:
        # 示例 1: 列出所有集群
        print("=== 列出所有集群 ===")
        clusters = client.list_clusters(page=1, page_size=20)
        print(f"响应: {clusters}\n")

        # 示例 2: 获取集群详情 (查询参数)
        print("=== 获取集群详情 (查询参数) ===")
        cluster = client.get_cluster("cluster-123")
        print(f"响应: {cluster}\n")

        # 示例 3: 列出命名空间 (查询参数)
        print("=== 列出命名空间 (查询参数) ===")
        namespaces = client.list_namespaces("cluster-123")
        print(f"响应: {namespaces}\n")

        # 示例 4: 获取命名空间详情 (查询参数)
        print("=== 获取命名空间详情 (查询参数) ===")
        namespace = client.get_namespace("cluster-123", "default")
        print(f"响应: {namespace}\n")

        # 示例 5: 列出 Pods (查询参数)
        print("=== 列出 Pods (查询参数) ===")
        pods = client.list_pods("cluster-123", "default", page=1, page_size=20)
        print(f"响应: {pods}\n")

        # 示例 6: 获取 Pod 详情 (查询参数)
        print("=== 获取 Pod 详情 (查询参数) ===")
        pod = client.get_pod("cluster-123", "default", "my-pod")
        print(f"响应: {pod}\n")

        # 示例 7: 获取 Pod 日志 (查询参数)
        print("=== 获取 Pod 日志 (查询参数) ===")
        logs = client.get_pod_logs("cluster-123", "default", "my-pod",
                                    container="app", tail_lines=100)
        print(f"响应: {logs}\n")

        # 示例 8: 列出 Deployments (查询参数)
        print("=== 列出 Deployments (查询参数) ===")
        deployments = client.list_deployments("cluster-123", "default")
        print(f"响应: {deployments}\n")

        # 示例 9: 获取 Deployment 详情 (查询参数)
        print("=== 获取 Deployment 详情 (查询参数) ===")
        deployment = client.get_deployment("cluster-123", "default", "my-deployment")
        print(f"响应: {deployment}\n")

        # 示例 10: 列出 Nodes (查询参数)
        print("=== 列出 Nodes (查询参数) ===")
        nodes = client.list_nodes("cluster-123")
        print(f"响应: {nodes}\n")

        # 示例 11: 测试 URL 编码 - 包含特殊字符的命名空间
        print("=== 测试 URL 编码 (kube-system) ===")
        kube_system = client.get_namespace("cluster-123", "kube-system")
        print(f"响应: {kube_system}\n")

    except requests.HTTPError as e:
        print(f"HTTP 错误: {e}")
        print(f"响应内容: {e.response.text}")
    except Exception as e:
        print(f"错误: {e}")


if __name__ == "__main__":
    main()
