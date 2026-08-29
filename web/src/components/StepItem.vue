<template>
  <div class="step">
    <li :class="step.status">
      <span class="step-name">* {{ step.name }}</span>
    </li>
    <pre v-if="step.errorMessage" class="error">{{ step.errorMessage }}</pre>
    <pre v-if="step.stackTrace" class="stack">{{ step.stackTrace }}</pre>
  </div>
</template>

<script setup>
defineProps({
  step: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
.step-name {
  display: inline-block;
}

.pass {
  color: #5e7d00;
}

.fail {
  color: #d80a16;
}

.progress {
  color: black;
}

.skip {
  color: #888;
}

li {
  list-style: none;
}

.error {
  margin: 4px 0 0 16px;
  color: #d80a16;
  font-size: 12px;
  white-space: pre-wrap;
  font-family: inherit;
}

.stack {
  margin: 2px 0 8px 16px;
  color: #8a3030;
  font-size: 11px;
  white-space: pre-wrap;
  font-family: inherit;
  opacity: 0.85;
}
</style>
