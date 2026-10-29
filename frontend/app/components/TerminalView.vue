<script setup lang="ts">
const host = ref<HTMLElement | null>(null)
const { start, dispose } = useTerminal(host)

onMounted(start)
onBeforeUnmount(dispose)
</script>

<template>
  <div class="crt">
    <div ref="host" class="term" />
    <div class="fx-scanlines" />
    <div class="fx-vignette" />
  </div>
</template>
