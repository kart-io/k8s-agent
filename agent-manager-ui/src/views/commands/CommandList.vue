<template>
  <div class="command-list">
    <a-card>
      <template #title>
        <div class="card-header">
          <span>命令执行</span>
          <a-space>
            <a-button type="primary" @click="showCreateModal">
              <template #icon>
                <plus-outlined />
              </template>
              新建命令
            </a-button>
            <a-button @click="loadCommands">
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
        :data="commands"
        :loading="loading"
        :row-config="{ isHover: true }"
        border
        stripe
        max-height="calc(100vh - 280px)"
      >
        <vxe-column type="seq" width="60" fixed="left" />
        <vxe-column field="id" title="命令 ID" width="280" fixed="left" />
        <vxe-column field="cluster_id" title="集群 ID" width="200" />
        <vxe-column field="type" title="类型" width="120" />
        <vxe-column field="action" title="操作" width="120" />
        <vxe-column field="status" title="状态" width="120">
          <template #default="{ row }">
            <a-tag :color="getStatusColor(row.status)">
              {{ getStatusText(row.status) }}
            </a-tag>
          </template>
        </vxe-column>
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
        <vxe-column title="操作" width="200" fixed="right">
          <template #default="{ row }">
            <a-space>
              <a-button type="link" size="small" @click="viewCommand(row)">
                详情
              </a-button>
              <a-button
                v-if="row.status === 'pending'"
                type="link"
                size="small"
                @click="executeCommand(row)"
              >
                执行
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

    <!-- Create Command Modal -->
    <a-modal
      v-model:open="createVisible"
      title="新建命令"
      width="600px"
      @ok="handleCreate"
    >
      <a-form :model="form" :label-col="{ span: 6 }">
        <a-form-item label="Agent ID" required>
          <a-input v-model:value="form.agent_id" placeholder="请输入 Agent ID" />
        </a-form-item>
        <a-form-item label="命令类型" required>
          <a-select v-model:value="form.type" placeholder="请选择命令类型">
            <a-select-option value="collect">收集信息</a-select-option>
            <a-select-option value="diagnose">诊断</a-select-option>
            <a-select-option value="restart">重启</a-select-option>
            <a-select-option value="custom">自定义</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="优先级">
          <a-select v-model:value="form.priority">
            <a-select-option :value="1">低</a-select-option>
            <a-select-option :value="5">中</a-select-option>
            <a-select-option :value="10">高</a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="超时时间(秒)">
          <a-input-number v-model:value="form.timeout" :min="1" :max="3600" />
        </a-form-item>
        <a-form-item label="参数">
          <a-textarea
            v-model:value="form.params"
            placeholder='请输入 JSON 格式参数，如: {"key": "value"}'
            :rows="4"
          />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- Command Detail Modal -->
    <a-modal
      v-model:open="detailVisible"
      title="命令详情"
      width="800px"
      :footer="null"
    >
      <a-descriptions v-if="currentCommand" :column="2" bordered>
        <a-descriptions-item label="命令 ID" :span="2">
          {{ currentCommand.id }}
        </a-descriptions-item>
        <a-descriptions-item label="集群 ID">
          {{ currentCommand.cluster_id }}
        </a-descriptions-item>
        <a-descriptions-item label="类型">
          {{ currentCommand.type }}
        </a-descriptions-item>
        <a-descriptions-item label="工具">
          {{ currentCommand.tool || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="操作">
          {{ currentCommand.action || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="状态">
          <a-tag :color="getStatusColor(currentCommand.status)">
            {{ getStatusText(currentCommand.status) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="命名空间">
          {{ currentCommand.namespace || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="超时时间">
          {{ currentCommand.timeout ? (currentCommand.timeout / 1000000000) + 's' : '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="发起者">
          {{ currentCommand.issued_by || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="创建时间">
          {{ formatTime(currentCommand.created_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="更新时间">
          {{ formatTime(currentCommand.updated_at) }}
        </a-descriptions-item>
        <a-descriptions-item label="关联 ID">
          {{ currentCommand.correlation_id || '-' }}
        </a-descriptions-item>
        <a-descriptions-item label="参数" :span="2">
          <a-tag v-for="(arg, index) in currentCommand.args" :key="index" style="margin: 2px">
            {{ arg }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item label="元数据" :span="2">
          <pre>{{ JSON.stringify(currentCommand.metadata, null, 2) }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import { getCommands, createCommand as createCommandApi, executeCommand as executeCommandApi } from '@/api/command'
import dayjs from 'dayjs'

const tableRef = ref()
const loading = ref(false)
const commands = ref([])
const createVisible = ref(false)
const detailVisible = ref(false)
const currentCommand = ref(null)

const form = reactive({
  agent_id: '',
  type: 'collect',
  priority: 5,
  timeout: 300,
  params: ''
})

const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0
})

const getStatusColor = (status) => {
  const colors = {
    pending: 'default',
    executing: 'processing',
    completed: 'success',
    failed: 'error',
    timeout: 'warning'
  }
  return colors[status] || 'default'
}

const getStatusText = (status) => {
  const texts = {
    pending: '待执行',
    executing: '执行中',
    completed: '已完成',
    failed: '失败',
    timeout: '超时'
  }
  return texts[status] || status
}

const formatTime = (time) => {
  return time ? dayjs(time).format('YYYY-MM-DD HH:mm:ss') : '-'
}

const loadCommands = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize
    }

    const res = await getCommands(params)
    commands.value = res.commands || []
    pagination.value.total = res.count || 0
  } catch (error) {
    message.error('加载命令列表失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = () => {
  loadCommands()
}

const showCreateModal = () => {
  Object.assign(form, {
    agent_id: '',
    type: 'collect',
    priority: 5,
    timeout: 300,
    params: ''
  })
  createVisible.value = true
}

const handleCreate = async () => {
  try {
    let params = {}
    if (form.params) {
      params = JSON.parse(form.params)
    }

    await createCommandApi({
      agent_id: form.agent_id,
      type: form.type,
      priority: form.priority,
      timeout: form.timeout,
      params
    })

    message.success('创建成功')
    createVisible.value = false
    loadCommands()
  } catch (error) {
    message.error('创建失败')
  }
}

const viewCommand = (row) => {
  currentCommand.value = row
  detailVisible.value = true
}

const executeCommand = async (row) => {
  try {
    await executeCommandApi(row.id)
    message.success('执行成功')
    loadCommands()
  } catch (error) {
    message.error('执行失败')
  }
}

onMounted(() => {
  loadCommands()
  // Auto refresh every 5 seconds
  setInterval(loadCommands, 5000)
})
</script>

<style lang="scss" scoped>
.command-list {
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
