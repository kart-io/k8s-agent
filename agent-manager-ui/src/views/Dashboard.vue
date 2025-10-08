<template>
  <div class="dashboard">
    <a-row :gutter="16">
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="在线 Agent"
            :value="stats.onlineAgents"
            :value-style="{ color: '#3f8600' }"
          >
            <template #prefix>
              <cluster-outlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic
            title="离线 Agent"
            :value="stats.offlineAgents"
            :value-style="{ color: '#cf1322' }"
          >
            <template #prefix>
              <disconnect-outlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="今日事件" :value="stats.todayEvents">
            <template #prefix>
              <alert-outlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
      <a-col :span="6">
        <a-card>
          <a-statistic title="执行中命令" :value="stats.runningCommands">
            <template #prefix>
              <code-outlined />
            </template>
          </a-statistic>
        </a-card>
      </a-col>
    </a-row>

    <a-row :gutter="16" class="mt-16">
      <a-col :span="12">
        <a-card title="最近事件" :bordered="false">
          <a-list
            :data-source="recentEvents"
            :loading="loading"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <a-tag :color="getEventColor(item.severity)">
                      {{ item.severity }}
                    </a-tag>
                    {{ item.type }}
                  </template>
                  <template #description>
                    {{ item.namespace }}/{{ item.labels?.name || '-' }} -
                    {{ formatTime(item.timestamp) }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>

      <a-col :span="12">
        <a-card title="Agent 状态" :bordered="false">
          <a-list
            :data-source="recentAgents"
            :loading="loading"
            item-layout="horizontal"
          >
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <a-badge
                      :status="item.status === 'online' ? 'success' : 'error'"
                      :text="item.id"
                    />
                  </template>
                  <template #description>
                    集群: {{ item.cluster_id }} - 最后心跳:
                    {{ formatTime(item.last_heartbeat) }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import {
  ClusterOutlined,
  DisconnectOutlined,
  AlertOutlined,
  CodeOutlined
} from '@ant-design/icons-vue'
import { getAgents, getAgentStats } from '@/api/agent'
import { getEvents, getEventStats } from '@/api/event'
import dayjs from 'dayjs'

const loading = ref(false)
const stats = ref({
  onlineAgents: 0,
  offlineAgents: 0,
  todayEvents: 0,
  runningCommands: 0
})
const recentEvents = ref([])
const recentAgents = ref([])

const getEventColor = (severity) => {
  const colors = {
    critical: 'red',
    warning: 'orange',
    info: 'blue',
    normal: 'green'
  }
  return colors[severity] || 'default'
}

const formatTime = (time) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const loadData = async () => {
  loading.value = true
  try {
    const [agentStatsRes, eventStatsRes, eventsRes, agentsRes] = await Promise.all([
      getAgentStats(),
      getEventStats(),
      getEvents({ page: 1, page_size: 5 }),
      getAgents({ page: 1, page_size: 5 })
    ])

    stats.value = {
      onlineAgents: agentStatsRes.online || 0,
      offlineAgents: agentStatsRes.offline || 0,
      todayEvents: eventStatsRes.today || 0,
      runningCommands: 0
    }

    recentEvents.value = eventsRes.events || []
    recentAgents.value = agentsRes.agents || []
  } catch (error) {
    console.error('Failed to load dashboard data:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
  // Auto refresh every 30 seconds
  setInterval(loadData, 30000)
})
</script>

<style lang="scss" scoped>
.dashboard {
  padding: 16px;
}
</style>
