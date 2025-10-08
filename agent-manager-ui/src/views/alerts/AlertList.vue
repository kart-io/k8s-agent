<template>
  <div class="alert-list">
    <a-card>
      <template #title>
        <div class="card-header">
          <span>告警规则</span>
          <a-space>
            <a-button type="primary" @click="showCreateModal">
              <template #icon>
                <plus-outlined />
              </template>
              新建规则
            </a-button>
            <a-button @click="loadAlerts">
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
        :data="alerts"
        :loading="loading"
        :row-config="{ isHover: true }"
        border
        stripe
        height="600"
      >
        <vxe-column field="id" title="规则 ID" width="280" fixed="left" />
        <vxe-column field="name" title="规则名称" width="200" />
        <vxe-column field="severity" title="严重程度" width="120">
          <template #default="{ row }">
            <a-tag :color="getSeverityColor(row.severity)">
              {{ row.severity }}
            </a-tag>
          </template>
        </vxe-column>
        <vxe-column field="enabled" title="状态" width="100">
          <template #default="{ row }">
            <a-switch
              v-model:checked="row.enabled"
              @change="toggleAlert(row)"
            />
          </template>
        </vxe-column>
        <vxe-column field="event_type" title="事件类型" width="150" />
        <vxe-column field="cluster_id" title="集群" width="150" />
        <vxe-column field="threshold" title="阈值" width="100" />
        <vxe-column field="created_at" title="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </vxe-column>
        <vxe-column title="操作" width="220" fixed="right">
          <template #default="{ row }">
            <a-space>
              <a-button type="link" size="small" @click="viewAlert(row)">
                详情
              </a-button>
              <a-button type="link" size="small" @click="editAlert(row)">
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

    <!-- Create/Edit Alert Modal -->
    <a-modal
      v-model:open="formVisible"
      :title="isEdit ? '编辑告警规则' : '新建告警规则'"
      width="700px"
      @ok="handleSubmit"
    >
      <a-form :model="form" :label-col="{ span: 6 }">
        <a-form-item label="规则名称" required>
          <a-input v-model:value="form.name" placeholder="请输入规则名称" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model:value="form.description"
            placeholder="请输入描述"
            :rows="2"
          />
        </a-form-item>
        <a-form-item label="严重程度" required>
          <a-select v-model:value="form.severity">
            <a-select-option value="critical">Critical</a-select-option>
            <a-select-option value="warning">Warning</a-select-option>
            <a-select-option value="info">Info</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="事件类型" required>
          <a-select v-model:value="form.event_type">
            <a-select-option value="pod.created">Pod Created</a-select-option>
            <a-select-option value="pod.deleted">Pod Deleted</a-select-option>
            <a-select-option value="pod.failed">Pod Failed</a-select-option>
            <a-select-option value="node.notready">Node NotReady</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="集群">
          <a-input
            v-model:value="form.cluster_id"
            placeholder="留空表示所有集群"
          />
        </a-form-item>
        <a-form-item label="命名空间">
          <a-input
            v-model:value="form.namespace"
            placeholder="留空表示所有命名空间"
          />
        </a-form-item>
        <a-form-item label="阈值">
          <a-input-number v-model:value="form.threshold" :min="1" />
        </a-form-item>
        <a-form-item label="时间窗口(分钟)">
          <a-input-number v-model:value="form.time_window" :min="1" />
        </a-form-item>
        <a-form-item label="通知渠道">
          <a-checkbox-group v-model:value="form.notification_channels">
            <a-checkbox value="email">邮件</a-checkbox>
            <a-checkbox value="webhook">Webhook</a-checkbox>
            <a-checkbox value="slack">Slack</a-checkbox>
          </a-checkbox-group>
        </a-form-item>
        <a-form-item label="条件">
          <a-textarea
            v-model:value="form.conditions"
            placeholder='请输入 JSON 格式条件'
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="启用规则">
          <a-switch v-model:checked="form.enabled" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Alert Detail Modal -->
    <a-modal
      v-model:open="detailVisible"
      title="告警规则详情"
      width="800px"
      :footer="null"
    >
      <a-descriptions v-if="currentAlert" :column="2" bordered>
        <a-descriptions-item label="规则 ID" :span="2">
          {{ currentAlert.id }}
        </a-descriptions-item>
        <a-descriptions-item label="规则名称">
          {{ currentAlert.name }}
        </a-descriptions-item>
        <a-descriptions-item label="严重程度">
          <a-tag :color="getSeverityColor(currentAlert.severity)">
            {{ currentAlert.severity }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="事件类型">
          {{ currentAlert.event_type }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-badge
            :status="currentAlert.enabled ? 'success' : 'default'"
            :text="currentAlert.enabled ? '已启用' : '已禁用'"
          />
        </a-descriptions-item>
        <a-descriptions-item label="集群">
          {{ currentAlert.cluster_id || '所有集群' }}
        </a-descriptions-item>
        <a-descriptions-item label="命名空间">
          {{ currentAlert.namespace || '所有命名空间' }}
        </a-descriptions-item>
        <a-descriptions-item label="阈值">
          {{ currentAlert.threshold }}
        </a-descriptions-item>
        <a-descriptions-item label="时间窗口">
          {{ currentAlert.time_window }} 分钟
        </a-descriptions-item>
        <a-descriptions-item label="描述" :span="2">
          {{ currentAlert.description }}
        </a-descriptions-item>
        <a-descriptions-item label="通知渠道" :span="2">
          <a-tag
            v-for="channel in currentAlert.notification_channels"
            :key="channel"
          >
            {{ channel }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(currentAlert.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatTime(currentAlert.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="条件" :span="2">
          <pre>{{ JSON.stringify(currentAlert.conditions, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { getAlerts, createAlert, updateAlert, deleteAlert } from '@/api/alert'
import dayjs from 'dayjs'

const tableRef = ref()
const loading = ref(false)
const alerts = ref([])
const formVisible = ref(false)
const detailVisible = ref(false)
const currentAlert = ref(null)
const isEdit = ref(false)

const form = reactive({
  name: '',
  description: '',
  severity: 'warning',
  event_type: 'pod.failed',
  cluster_id: '',
  namespace: '',
  threshold: 1,
  time_window: 5,
  notification_channels: [],
  conditions: '',
  enabled: true
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
    info: 'blue'
  }
  return colors[severity] || 'default'
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const loadAlerts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    }

    const res = await getAlerts(params)
    alerts.value = res.alerts || []
    pagination.value.total = res.count || 0
  } catch (error) {
    message.error('加载告警规则失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = () => {
  loadAlerts()
}

const showCreateModal = () => {
  isEdit.value = false
  Object.assign(form, {
    name: '',
    description: '',
    severity: 'warning',
    event_type: 'pod.failed',
    cluster_id: '',
    namespace: '',
    threshold: 1,
    time_window: 5,
    notification_channels: [],
    conditions: '',
    enabled: true
  })
  formVisible.value = true
}

const editAlert = (row) => {
  isEdit.value = true
  Object.assign(form, {
    name: row.name,
    description: row.description || '',
    severity: row.severity,
    event_type: row.event_type,
    cluster_id: row.cluster_id || '',
    namespace: row.namespace || '',
    threshold: row.threshold,
    time_window: row.time_window,
    notification_channels: row.notification_channels || [],
    conditions: row.conditions ? JSON.stringify(row.conditions, null, 2) : '',
    enabled: row.enabled
  })
  currentAlert.value = row
  formVisible.value = true
}

const handleSubmit = async () => {
  try {
    let conditions = {}
    if (form.conditions) {
      conditions = JSON.parse(form.conditions)
    }

    const data = {
      name: form.name,
      description: form.description,
      severity: form.severity,
      event_type: form.event_type,
      cluster_id: form.cluster_id,
      namespace: form.namespace,
      threshold: form.threshold,
      time_window: form.time_window,
      notification_channels: form.notification_channels,
      conditions,
      enabled: form.enabled
    }

    if (isEdit.value) {
      await updateAlert(currentAlert.value.id, data)
      message.success('更新成功')
    } else {
      await createAlert(data)
      message.success('创建成功')
    }

    formVisible.value = false
    loadAlerts()
  } catch (error) {
    message.error(isEdit.value ? '更新失败' : '创建失败')
  }
}

const viewAlert = (row) => {
  currentAlert.value = row
  detailVisible.value = true
}

const toggleAlert = async (row) => {
  try {
    await updateAlert(row.id, { enabled: row.enabled })
    message.success(row.enabled ? '已启用' : '已禁用')
  } catch (error) {
    message.error('操作失败')
    row.enabled = !row.enabled
  }
}

const handleDelete = (row) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除告警规则 "${row.name}" 吗？`,
    onOk: async () => {
      try {
        await deleteAlert(row.id)
        message.success('删除成功')
        loadAlerts()
      } catch (error) {
        message.error('删除失败')
      }
    }
  })
}

onMounted(() => {
  loadAlerts()
})
</script>

<style lang="scss" scoped>
.alert-list {
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
