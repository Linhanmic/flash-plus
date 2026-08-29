<template>
  <div class="menu" :class="{ 'fail-bg': failed, 'pass-bg': passed && finished }">
    <div class="stats-container">
      <span class="entity">Specifications: </span>
      <span class="stat">{{ stats.specPassed }} passed, </span>
      <span class="stat">{{ stats.specFailed }} failed, </span>
      <span class="stat">{{ stats.specProgress }} running</span>
      <span> | </span>
      <span class="entity">Scenarios: </span>
      <span class="stat">{{ stats.scenarioPassed }} passed, </span>
      <span class="stat">{{ stats.scenarioFailed }} failed, </span>
      <span class="stat">{{ stats.scenarioProgress }} running</span>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useEventStore } from '../stores/eventStore'

const store = useEventStore()

const stats = computed(() => store.stats)
const finished = computed(() => store.isFinished)
const failed = computed(() => store.finalStatus === 'fail')
const passed = computed(() => store.finalStatus === 'pass')
</script>

<style scoped>
.menu {
  background: gray;
  padding: 8px 10px;
  color: #e2e2e2;
}

.fail-bg {
  background: #e73e48;
}

.pass-bg {
  background: #5e7d00;
}

.entity {
  color: white;
}

.stat {
  font-style: italic;
  font-weight: 100;
}
</style>
