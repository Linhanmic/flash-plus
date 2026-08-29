import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

let nextId = 1
const uid = () => nextId++

function lastOpen(items, name) {
  for (let i = items.length - 1; i >= 0; i--) {
    if (items[i].name === name && items[i].status === 'progress') {
      return items[i]
    }
  }
  return null
}

export const useEventStore = defineStore('event', () => {
  const events = ref([])
  const specs = ref({})
  const isFinished = ref(false)
  const finalStatus = ref('pass')
  const projectName = ref('')
  const executionTime = ref(0)
  const environment = ref('')
  const specsSkipped = ref(0)

  const specList = computed(() => Object.values(specs.value))

  const stats = computed(() => {
    const result = {
      specPassed: 0,
      specFailed: 0,
      specProgress: 0,
      specSkipped: specsSkipped.value,
      scenarioPassed: 0,
      scenarioFailed: 0,
      scenarioProgress: 0
    }
    for (const spec of specList.value) {
      if (spec.status === 'progress') result.specProgress++
      else if (spec.status === 'pass') result.specPassed++
      else if (spec.status === 'fail') result.specFailed++
      else if (spec.status === 'skip') result.specSkipped++

      for (const scenario of spec.scenarios || []) {
        if (scenario.status === 'progress') result.scenarioProgress++
        else if (scenario.status === 'pass') result.scenarioPassed++
        else if (scenario.status === 'fail') result.scenarioFailed++
      }
    }
    return result
  })

  function addEvent(event) {
    if (!event || !event.type) return
    events.value.push(event)

    switch (event.type) {
      case 'suite':
        if (event.projectName) projectName.value = event.projectName
        else if (event.name) projectName.value = event.name
        break

      case 'spec': {
        const key = event.fileName || event.name || `spec-${nextId}`
        if (event.status === 'progress' || !specs.value[key]) {
          if (!specs.value[key]) {
            specs.value[key] = {
              id: uid(),
              ...event,
              scenarios: []
            }
          } else {
            specs.value[key].status = event.status
            if (event.name) specs.value[key].name = event.name
          }
        } else {
          specs.value[key].status = event.status
        }
        break
      }

      case 'scenario': {
        const spec = specs.value[event.specFileName]
        if (!spec) break
        if (event.status === 'progress') {
          spec.scenarios.push({
            id: uid(),
            ...event,
            steps: []
          })
        } else {
          const open = lastOpen(spec.scenarios, event.name)
          if (open) {
            open.status = event.status
          } else {
            spec.scenarios.push({ id: uid(), ...event, steps: [] })
          }
        }
        break
      }

      case 'step': {
        const spec = specs.value[event.specFileName]
        if (!spec) break
        let scenario = lastOpen(spec.scenarios, event.scenarioName)
        if (!scenario) {
          scenario = spec.scenarios.find((s) => s.name === event.scenarioName)
        }
        if (!scenario) break
        if (event.status === 'progress') {
          scenario.steps.push({ id: uid(), ...event })
        } else {
          const open = lastOpen(scenario.steps, event.name)
          if (open) {
            open.status = event.status
            open.errorMessage = event.errorMessage
            open.stackTrace = event.stackTrace
          } else {
            scenario.steps.push({ id: uid(), ...event })
          }
        }
        break
      }

      case 'end':
        isFinished.value = true
        finalStatus.value = event.status || 'pass'
        if (event.projectName) projectName.value = event.projectName
        if (event.executionTime) executionTime.value = event.executionTime
        if (event.environment) environment.value = event.environment
        if (event.specsSkipped) specsSkipped.value = event.specsSkipped
        break
    }
  }

  function addEvents(list) {
    for (const ev of list) addEvent(ev)
  }

  function reset() {
    events.value = []
    specs.value = {}
    isFinished.value = false
    finalStatus.value = 'pass'
    projectName.value = ''
    executionTime.value = 0
    environment.value = ''
    specsSkipped.value = 0
    nextId = 1
  }

  return {
    events,
    specs,
    specList,
    stats,
    isFinished,
    finalStatus,
    projectName,
    executionTime,
    environment,
    specsSkipped,
    addEvent,
    addEvents,
    reset
  }
})
