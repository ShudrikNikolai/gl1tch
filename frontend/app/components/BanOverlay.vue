<script setup lang="ts">
import { useSessionStore } from '~/stores/session'

const session = useSessionStore()
// поменяй ссылку на свою
const RICKROLL = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'
</script>

<template>
  <div v-if="session.status === 'banned'" class="ban" role="alert">
    <a :href="RICKROLL" target="_blank" rel="noopener noreferrer">you fucked up</a>
  </div>
</template>
