<template>
  <div class="scenario">
    <li :class="scenario.status">
      <span class="scenario-name">## {{ scenario.name }}</span>
    </li>
    <ul>
      <StepItem
        v-for="step in scenario.steps"
        :key="step.id"
        :step="step"
      />
    </ul>
  </div>
</template>

<script setup>
import StepItem from './StepItem.vue'

defineProps({
  scenario: {
    type: Object,
    required: true
  }
})
</script>

<style scoped>
.scenario-name {
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

ul {
  margin-left: 20px;
  padding-left: 0;
}

li {
  list-style: none;
}
</style>
