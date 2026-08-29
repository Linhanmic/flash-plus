<template>
  <div class="step">
    <li :class="step.status">
      <span class="step-name">* {{ step.name }}</span>
    </li>
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
</style>
