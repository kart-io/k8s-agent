<template>
  <div class="agent-list">
    <a-card>
      <template #title>
        <div class="card-header">
          <span>Agent 列表</span>
          <a-space>
            <a-input-search
              v-model:value="searchText"
              placeholder="搜索 Agent ID 或集群"
              style="width: 250px"
              @search="handleSearch"
            />
            <a-button type="primary" @click="loadAgents">
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
        :data="agents"
        :loading="loading"
        :row-config="{ isHover: true }"
        border
        stripe
        max-height="calc(100vh - 280px)"
      >
        <vxe-column field="id" title="Agent ID" width="200" fixed="left" />
        <vxe-column field="cluster_id" title="集群 ID" width="180" />
        <vxe-column field="cluster_name" title="集群名称" width="150">
          <template #default="{ row }">
            {{ row.cluster_name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="connection_info.local_ip" title="服务 IP" width="150">
          <template #default="{ row }">
            {{ row.connection_info?.local_ip || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="version" title="版本" width="120" />
        <vxe-column field="status" title="状态" width="100">
          <template #default="{ row }">
            <a-badge
              :status="row.status === 'online' ? 'success' : 'error'"
              :text="row.status === 'online' ? '在线' : '离线'"
            />
          </template>
        </vxe-column>
        <vxe-column field="last_heartbeat" title="最后心跳" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_heartbeat) }}
          </template>
        </vxe-column>
        <vxe-column field="registered_at" title="注册时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.registered_at) }}
          </template>
        </vxe-column>
        <vxe-column title="操作" width="180" fixed="right">
          <template #default="{ row }">
            <a-space>
              <a-button type="link" size="small" @click="viewAgent(row)">
                详情
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

    <!-- Agent Detail Modal -->
    <a-modal
      v-model:open="detailVisible"
      title="Agent 详情"
      width="800px"
      :footer="null"
    >
      <a-descriptions v-if="currentAgent" :column="2" bordered>
        <a-descriptions-item label="Agent ID" :span="2">
          {{ currentAgent.id }}
        </a-descriptions-item>
        <a-descriptions-item label="集群 ID">
          {{ currentAgent.cluster_id }}
        </a-descriptions-item>
        <a-descriptions-item label="集群名称">
          {{ currentAgent.cluster_name || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="版本">
          {{ currentAgent.version }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-badge
            :status="currentAgent.status === 'online' ? 'success' : 'error'"
            :text="currentAgent.status === 'online' ? '在线' : '离线'"
          />
        </a-descriptions-item>
        <a-descriptions-item label="最后心跳">
          {{ formatTime(currentAgent.last_heartbeat) }}
        </a-descriptions-item>
        <a-descriptions-item label="注册时间">
          {{ formatTime(currentAgent.registered_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatTime(currentAgent.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="服务地址">
          {{ currentAgent.connection_info?.service_address || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="本地 IP">
          {{ currentAgent.connection_info?.local_ip || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="NATS 端点">
          {{ currentAgent.connection_info?.endpoint || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="连接时间">
          {{ formatTime(currentAgent.connection_info?.connected_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="能力" :span="2">
          <a-tag v-for="cap in currentAgent.capabilities" :key="cap">{{ cap }}</a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="元数据" :span="2">
          <pre>{{ JSON.stringify(currentAgent.metadata, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { getAgents, deleteAgent } from '@/api/agent'
import dayjs from 'dayjs'

const tableRef = ref()
const loading = ref(false)
const agents = ref([])
const searchText = ref('')
const detailVisible = ref(false)
const currentAgent = ref(null)

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0
})

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const loadAgents = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    }
    if (searchText.value) {
      params.search = searchText.value
    }

    const res = await getAgents(params)
    agents.value = res.agents || []
    pagination.value.total = res.count || 0
  } catch (error) {
    message.error('加载 Agent 列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.value.current = 1
  loadAgents()
}

const handlePageChange = () => {
  loadAgents()
}

const viewAgent = (row) => {
  currentAgent.value = row
  detailVisible.value = true
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除 Agent "${row.id}" 吗？`,
    onOk: async () => {
      try {
        await deleteAgent(row.id)
        message.success('删除成功')
        loadAgents()
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

onMounted(() => {
  loadAgents()
  // Auto refresh every 30 seconds
  setInterval(loadAgents, 30000)
})
</script>

<style lang="scss" scoped>
.agent-list {
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
