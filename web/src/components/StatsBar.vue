<template>
  <div class="menu" :class="{ 'fail-bg': store.failed, 'pass-bg': store.passed && store.isFinished }">
    <div class="toggle-all">
      <input
        id="collapse"
        type="button"
        :value="store.allExpanded ? 'Hide all' : 'Show all'"
        @click="store.toggleAll()"
      />
    </div>
    <div class="stats-container">
      <span class="entity">Specifications: </span>
      <span class="stat">{{ store.stats.specPassed }} passed, </span>
      <span class="stat">{{ store.stats.specFailed }} failed, </span>
      <span class="stat" v-if="store.stats.specSkipped">{{ store.stats.specSkipped }} skipped, </span>
      <span class="stat">{{ store.stats.specProgress }} running</span>
      <span> | </span>
      <span class="entity">Scenarios: </span>
      <span class="stat">{{ store.stats.scenarioPassed }} passed, </span>
      <span class="stat">{{ store.stats.scenarioFailed }} failed, </span>
      <span class="stat">{{ store.stats.scenarioProgress }} running</span>
      <span v-if="store.environment" class="env"> · {{ store.environment }}</span>
    </div>
  </div>
</template>

<script setup>
import { useReportStore } from '../stores/report'

const store = useReportStore()
</script>

<style scoped>
.menu {
  background: gray;
  padding: 8px 10px;
  color: #e2e2e2;
  overflow: hidden;
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

.env {
  font-weight: 100;
  opacity: 0.85;
}

.toggle-all {
  float: right;
}

#collapse {
  cursor: pointer;
  color: #ececec;
  border: 0.5px solid #ececec;
  border-radius: 5px;
  background: transparent;
  min-width: 70px;
  font-family: inherit;
  font-size: 13px;
}

#collapse:focus {
  outline: 0;
}
</style>
