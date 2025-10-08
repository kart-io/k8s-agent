<template>
  <div class="event-list">
    <a-card>
      <template #title>
        <div class="card-header">
          <span>事件监控</span>
          <a-space>
            <a-select
              v-model:value="filters.severity"
              placeholder="严重程度"
              style="width: 120px"
              allow-clear
              @change="handleFilter"
            >
              <a-select-option value="critical">Critical</a-select-option>
              <a-select-option value="warning">Warning</a-select-option>
              <a-select-option value="info">Info</a-select-option>
              <a-select-option value="normal">Normal</a-select-option>
            </a-select>
            <a-select
              v-model:value="filters.type"
              placeholder="事件类型"
              style="width: 150px"
              allow-clear
              @change="handleFilter"
            >
              <a-select-option value="pod.created">Pod Created</a-select-option>
              <a-select-option value="pod.deleted">Pod Deleted</a-select-option>
              <a-select-option value="pod.failed">Pod Failed</a-select-option>
              <a-select-option value="node.notready">Node NotReady</a-select-option>
            </a-select>
            <a-input-search
              v-model:value="searchText"
              placeholder="搜索资源名称"
              style="width: 200px"
              @search="handleSearch"
            />
            <a-button type="primary" @click="loadEvents">
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
        :data="events"
        :loading="loading"
        :row-config="{ isHover: true }"
        border
        stripe
        max-height="calc(100vh - 320px)"
      >
        <vxe-column type="seq" width="60" fixed="left" />
        <vxe-column field="severity" title="严重程度" width="100" fixed="left">
          <template #default="{ row }">
            <a-tag :color="getSeverityColor(row.severity)">
              {{ row.severity }}
            </a-tag>
          </template>
        </vxe-column>
        <vxe-column field="type" title="事件类型" width="180" />
        <vxe-column field="cluster_id" title="集群" width="150" />
        <vxe-column field="labels.kind" title="资源类型" width="120">
          <template #default="{ row }">
            {{ row.labels?.kind || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="namespace" title="命名空间" width="150" />
        <vxe-column field="labels.name" title="资源名称" width="200">
          <template #default="{ row }">
            {{ row.labels?.name || '-' }}
          </template>
        </vxe-column>
        <vxe-column field="reason" title="原因" width="150" />
        <vxe-column field="timestamp" title="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.timestamp) }}
          </template>
        </vxe-column>
        <vxe-column title="操作" width="120" fixed="right">
          <template #default="{ row }">
            <a-button type="link" size="small" @click="viewEvent(row)">
              详情
            </a-button>
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

    <!-- Event Detail Modal -->
    <a-modal
      v-model:open="detailVisible"
      title="事件详情"
      width="800px"
      :footer="null"
    >
      <a-descriptions v-if="currentEvent" :column="2" bordered>
        <a-descriptions-item label="事件 ID" :span="2">
          {{ currentEvent.id }}
        </a-descriptions-item>
        <a-descriptions-item label="严重程度">
          <a-tag :color="getSeverityColor(currentEvent.severity)">
            {{ currentEvent.severity }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="事件类型">
          {{ currentEvent.type }}
        </a-descriptions-item>
        <a-descriptions-item label="集群">
          {{ currentEvent.cluster_id }}
        </a-descriptions-item>
        <a-descriptions-item label="资源类型">
          {{ currentEvent.labels?.kind || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="命名空间">
          {{ currentEvent.namespace || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="资源名称">
          {{ currentEvent.labels?.name || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="原因">
          {{ currentEvent.reason }}
        </a-descriptions-item>
        <a-descriptions-item label="消息" :span="2">
          {{ currentEvent.message }}
        </a-descriptions-item>
        <a-descriptions-item label="时间">
          {{ formatTime(currentEvent.timestamp) }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(currentEvent.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="元数据" :span="2">
          <pre>{{ JSON.stringify(currentEvent.metadata, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { ReloadOutlined } from '@ant-design/icons-vue'
import { getEvents } from '@/api/event'
import dayjs from 'dayjs'

const tableRef = ref()
const loading = ref(false)
const events = ref([])
const searchText = ref('')
const detailVisible = ref(false)
const currentEvent = ref(null)

const filters = ref({
  severity: undefined,
  type: undefined
})

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0
})

const getSeverityColor = (severity) => {
  const colors = {
    critical: 'red',
    warning: 'orange',
    info: 'blue',
    normal: 'green'
  }
  return colors[severity] || 'default'
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const loadEvents = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    }

    if (searchText.value) {
      params.search = searchText.value
    }
    if (filters.value.severity) {
      params.severity = filters.value.severity
    }
    if (filters.value.type) {
      params.type = filters.value.type
    }

    const res = await getEvents(params)
    events.value = res.events || []
    pagination.value.total = res.count || 0
  } catch (error) {
    message.error('加载事件列表失败')
  } finally {
    loading.value = false
  }
}

const handleFilter = () => {
  pagination.value.current = 1
  loadEvents()
}

const handleSearch = () => {
  pagination.value.current = 1
  loadEvents()
}

const handlePageChange = () => {
  loadEvents()
}

const viewEvent = (row) => {
  currentEvent.value = row
  detailVisible.value = true
}

onMounted(() => {
  loadEvents()
  // Auto refresh every 10 seconds
  setInterval(loadEvents, 10000)
})
</script>

<style lang="scss" scoped>
.event-list {
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
