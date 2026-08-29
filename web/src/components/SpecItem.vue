<template>
  <div class="spec">
    <li :class="spec.status" @click="expanded = !expanded">
      <span class="spec-name"># {{ spec.name }}</span>
      <span class="toggle">{{ expanded ? 'hide' : 'show' }}</span>
    </li>
    <ul v-show="expanded" class="scenarios">
      <ScenarioItem
        v-for="scenario in spec.scenarios"
        :key="scenario.id"
        :scenario="scenario"
      />
    </ul>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import ScenarioItem from './ScenarioItem.vue'

const props = defineProps({
  spec: {
    type: Object,
    required: true
  },
  forceExpanded: {
    type: Boolean,
    default: true
  }
})

const expanded = ref(true)

watch(
  () => props.forceExpanded,
  (val) => {
    expanded.value = val
  }
)
</script>

<style scoped>
.spec {
  margin-bottom: 0.2%;
  border-bottom: 1px solid #eaeaea;
}

.spec:hover {
  background: #f1f1f1;
}

.spec-name {
  display: inline-block;
}

.toggle {
  cursor: pointer;
  font-size: 10px;
  color: #9e9e9e;
  margin-left: 15px;
  padding: 0 5px;
  border: 0.5px solid #cccccc;
  display: none;
}

.spec:hover .toggle {
  display: inline;
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

.scenarios {
  margin-left: 20px;
  padding-left: 0;
}

li {
  list-style: none;
  cursor: pointer;
}
</style>
