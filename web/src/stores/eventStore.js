import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useEventStore = defineStore('event', () => {
  const events = ref([])
  const specs = ref({})
  const isFinished = ref(false)
  const finalStatus = ref('pass')

  // 统计数据
  const stats = computed(() => {
    const result = {
      specPassed: 0,
      specFailed: 0,
      specProgress: 0,
      scenarioPassed: 0,
      scenarioFailed: 0,
      scenarioProgress: 0
    }
    for (const spec of Object.values(specs.value)) {
      if (spec.status === 'progress') result.specProgress++
      else if (spec.status === 'pass') result.specPassed++
      else if (spec.status === 'fail') result.specFailed++

      for (const scenario of Object.values(spec.scenarios || {})) {
        if (scenario.status === 'progress') result.scenarioProgress++
        else if (scenario.status === 'pass') result.scenarioPassed++
        else if (scenario.status === 'fail') result.scenarioFailed++
      }
    }
    return result
  })

  // 处理新事件
  function addEvent(event) {
    events.value.push(event)

    switch (event.type) {
      case 'spec':
        if (!specs.value[event.fileName]) {
          specs.value[event.fileName] = {
            ...event,
            scenarios: {}
          }
        } else {
          specs.value[event.fileName].status = event.status
        }
        break

      case 'scenario':
        const spec = specs.value[event.specFileName]
        if (spec) {
          if (!spec.scenarios[event.name]) {
            spec.scenarios[event.name] = {
              ...event,
              steps: {}
            }
          } else {
            spec.scenarios[event.name].status = event.status
          }
        }
        break

      case 'step':
        const stepSpec = specs.value[event.specFileName]
        const scenario = stepSpec?.scenarios[event.scenarioName]
        if (scenario) {
          if (!scenario.steps[event.name]) {
            scenario.steps[event.name] = { ...event }
          } else {
            scenario.steps[event.name].status = event.status
          }
        }
        break

      case 'end':
        isFinished.value = true
        finalStatus.value = event.status
        break
    }
  }

  return { events, specs, stats, isFinished, finalStatus, addEvent }
})
