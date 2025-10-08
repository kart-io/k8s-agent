<template>
  <a-config-provider :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>

<script setup>
import { computed } from 'vue'
import { theme } from 'ant-design-vue'

const themeConfig = computed(() => ({
  algorithm: theme.defaultAlgorithm,
  token: {
    colorPrimary: '#1890ff',
    borderRadius: 6
  }
}))
</script>

<style>
#app {
  height: 100vh;
  overflow: hidden;
}
</style>
