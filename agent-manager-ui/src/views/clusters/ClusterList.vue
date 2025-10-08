<template>
  <div class="cluster-list">
    <a-card>
      <template #title>
        <div class="card-header">
          <span>集群管理</span>
          <a-space>
            <a-button type="primary" @click="showCreateModal">
              <template #icon>
                <plus-outlined />
              </template>
              添加集群
            </a-button>
            <a-button @click="loadClusters">
              <template #icon>
                <reload-outlined />
              </template>
              刷新
            </a-button>
          </a-space>
        </div>
      </template>

      <vxe-table
        ref="tableRef"
        :data="clusters"
        :loading="loading"
        :row-config="{ isHover: true }"
        border
        stripe
        max-height="calc(100vh - 280px)"
      >
        <vxe-column field="id" title="集群 ID" width="200" fixed="left" />
        <vxe-column field="name" title="集群名称" width="180" />
        <vxe-column field="region" title="区域" width="120">
          <template #default="{ row }">
            {{ row.region || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="version" title="K8s 版本" width="150" />
        <vxe-column field="api_server" title="API Server" width="280">
          <template #default="{ row }">
            {{ row.api_server || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="status" title="状态" width="100">
          <template #default="{ row }">
            <a-badge
              :status="row.status === 'active' ? 'success' : 'error'"
              :text="row.status === 'active' ? '正常' : '异常'"
            />
          </template>
        </vxe-column>
        <vxe-column field="node_count" title="节点数" width="100" />
        <vxe-column field="created_at" title="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </vxe-column>
        <vxe-column field="updated_at" title="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </vxe-column>
        <vxe-column title="操作" width="220" fixed="right">
          <template #default="{ row }">
            <a-space>
              <a-button type="link" size="small" @click="viewCluster(row)">
                详情
              </a-button>
              <a-button type="link" size="small" @click="editCluster(row)">
                编辑
              </a-button>
              <a-button
                type="link"
                size="small"
                danger
                @click="handleDelete(row)"
              >
                删除
              </a-button>
            </a-space>
          </template>
        </vxe-column>
      </vxe-table>

      <div class="pagination">
        <a-pagination
          v-model:current="pagination.current"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :show-total="(total) => `共 ${total} 条`"
          :show-size-changer="true"
          @change="handlePageChange"
        />
      </div>
    </a-card>

    <!-- Create/Edit Cluster Modal -->
    <a-modal
      v-model:open="formVisible"
      :title="isEdit ? '编辑集群' : '添加集群'"
      width="600px"
      @ok="handleSubmit"
    >
      <a-form :model="form" :label-col="{ span: 6 }">
        <a-form-item label="集群 ID" required>
          <a-input
            v-model:value="form.cluster_id"
            placeholder="请输入集群 ID"
            :disabled="isEdit"
          />
        </a-form-item>
        <a-form-item label="集群名称" required>
          <a-input v-model:value="form.name" placeholder="请输入集群名称" />
        </a-form-item>
        <a-form-item label="区域">
          <a-input v-model:value="form.region" placeholder="如: us-west-1" />
        </a-form-item>
        <a-form-item label="K8s 版本">
          <a-input v-model:value="form.version" placeholder="如: 1.28.0" />
        </a-form-item>
        <a-form-item label="API Server">
          <a-input
            v-model:value="form.api_server"
            placeholder="https://api.cluster.example.com"
          />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model:value="form.description"
            placeholder="请输入描述"
            :rows="3"
          />
        </a-form-item>
        <a-form-item label="配置">
          <a-textarea
            v-model:value="form.config"
            placeholder='请输入 JSON 格式配置'
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Cluster Detail Modal -->
    <a-modal
      v-model:open="detailVisible"
      title="集群详情"
      width="800px"
      :footer="null"
    >
      <a-descriptions v-if="currentCluster" :column="2" bordered>
        <a-descriptions-item label="集群 ID" :span="2">
          {{ currentCluster.id }}
        </a-descriptions-item>
        <a-descriptions-item label="集群名称">
          {{ currentCluster.name }}
        </a-descriptions-item>
        <a-descriptions-item label="区域">
          {{ currentCluster.region || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="提供商">
          {{ currentCluster.provider || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="环境">
          {{ currentCluster.environment || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="K8s 版本">
          {{ currentCluster.version || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="API Server">
          {{ currentCluster.api_server || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-badge
            :status="currentCluster.status === 'active' ? 'success' : 'error'"
            :text="currentCluster.status === 'active' ? '正常' : '异常'"
          />
        </a-descriptions-item>
        <a-descriptions-item label="健康状态">
          <a-badge
            :status="currentCluster.health === 'healthy' ? 'success' : 'warning'"
            :text="currentCluster.health || '-'"
          />
        </a-descriptions-item>
        <a-descriptions-item label="Agent 数">
          {{ currentCluster.agent_count }}
        </a-descriptions-item>
        <a-descriptions-item label="节点数">
          {{ currentCluster.node_count }}
        </a-descriptions-item>
        <a-descriptions-item label="Pod 数">
          {{ currentCluster.pod_count }}
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">
          {{ currentCluster.description || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(currentCluster.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatTime(currentCluster.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="元数据" :span="2">
          <pre>{{ JSON.stringify(currentCluster.metadata, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import {
  getClusters,
  createCluster,
  updateCluster,
  deleteCluster
} from '@/api/cluster'
import dayjs from 'dayjs'

const tableRef = ref()
const loading = ref(false)
const clusters = ref([])
const formVisible = ref(false)
const detailVisible = ref(false)
const currentCluster = ref(null)
const isEdit = ref(false)

const form = reactive({
  cluster_id: '',
  name: '',
  region: '',
  version: '',
  api_server: '',
  description: '',
  config: ''
})

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0
})

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const loadClusters = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    }

    const res = await getClusters(params)
    clusters.value = res.clusters || []
    pagination.value.total = res.count || 0
  } catch (error) {
    message.error('加载集群列表失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = () => {
  loadClusters()
}

const showCreateModal = () => {
  isEdit.value = false
  Object.assign(form, {
    cluster_id: '',
    name: '',
    region: '',
    version: '',
    api_server: '',
    description: '',
    config: ''
  })
  formVisible.value = true
}

const editCluster = (row) => {
  isEdit.value = true
  Object.assign(form, {
    cluster_id: row.id,
    name: row.name,
    region: row.region || '',
    version: row.version || '',
    api_server: row.api_server || '',
    description: row.description || '',
    config: row.metadata ? JSON.stringify(row.metadata, null, 2) : ''
  })
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    let config = {}
    if (form.config) {
      config = JSON.parse(form.config)
    }

    const data = {
      cluster_id: form.cluster_id,
      name: form.name,
      region: form.region,
      version: form.version,
      api_server: form.api_server,
      description: form.description,
      config
    }

    if (isEdit.value) {
      await updateCluster(form.cluster_id, data)
      message.success('更新成功')
    } else {
      await createCluster(data)
      message.success('创建成功')
    }

    formVisible.value = false
    loadClusters()
  } catch (error) {
    message.error(isEdit.value ? '更新失败' : '创建失败')
  }
}

const viewCluster = (row) => {
  currentCluster.value = row
  detailVisible.value = true
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除集群 "${row.name}" 吗？`,
    onOk: async () => {
      try {
        await deleteCluster(row.id)
        message.success('删除成功')
        loadClusters()
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

onMounted(() => {
  loadClusters()
})
</script>

<style lang="scss" scoped>
.cluster-list {
  padding: 16px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .pagination {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }

  pre {
    background: #f5f5f5;
    padding: 8px;
    border-radius: 4px;
    max-height: 300px;
    overflow: auto;
  }
}
</style>
