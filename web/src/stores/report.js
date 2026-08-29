import { defineStore } from 'pinia'
import { computed, ref, watch } from 'vue'

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

function collectExpandableIds(nodes, acc = []) {
  for (const node of nodes || []) {
    if (node.children?.length) {
      acc.push(String(node.id))
      collectExpandableIds(node.children, acc)
    }
  }
  return acc
}

function formatTimestamp(date = new Date()) {
  const pad = (n) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

export const useReportStore = defineStore('report', () => {
  const events = ref([])
  const specs = ref({})
  const isFinished = ref(false)
  const finalStatus = ref('pass')
  const projectName = ref('')
  const executionTime = ref(0)
  const environment = ref('')
  const specsSkipped = ref(0)
  const timestamp = ref('')

  const isConnected = ref(false)
  const allExpanded = ref(true)
  const expandedKeys = ref([])

  let socket = null
  let reconnectTimer = null
  let closed = false

  const specList = computed(() => Object.values(specs.value))

  const treeData = computed(() =>
    specList.value.map((spec) => ({
      id: `spec-${spec.id}`,
      name: spec.name,
      type: 'spec',
      status: spec.status || 'progress',
      children: (spec.scenarios || []).map((scenario) => ({
        id: `scn-${scenario.id}`,
        name: scenario.name,
        type: 'scenario',
        status: scenario.status || 'progress',
        children: (scenario.steps || []).map((step) => {
          const node = {
            id: `step-${step.id}`,
            name: step.name,
            type: 'step',
            status: step.status || 'progress',
            errorMessage: step.errorMessage,
            stackTrace: step.stackTrace
          }
          if (step.errorMessage || step.stackTrace) {
            node.children = [
              {
                id: `err-${step.id}`,
                name: step.errorMessage || 'Step failed',
                type: 'error',
                status: 'fail',
                stackTrace: step.stackTrace
              }
            ]
          }
          return node
        })
      }))
    }))
  )

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

  const durationText = computed(() => {
    const ms = executionTime.value
    if (!ms) return ''
    if (ms < 1000) return `${ms} ms`
    return `${(ms / 1000).toFixed(2)} s`
  })

  const projectLabel = computed(() => (projectName.value ? ` · ${projectName.value}` : ''))
  const failed = computed(() => finalStatus.value === 'fail')
  const passed = computed(() => finalStatus.value === 'pass')

  function syncExpandedKeys() {
    expandedKeys.value = allExpanded.value ? collectExpandableIds(treeData.value) : []
  }

  watch(treeData, () => {
    if (allExpanded.value) {
      syncExpandedKeys()
    }
  })

  function toggleAll() {
    allExpanded.value = !allExpanded.value
    syncExpandedKeys()
  }

  function onExpandChange(row, expanded) {
    if (typeof expanded !== 'boolean' || !row) return
    const id = String(row.id)
    if (expanded) {
      if (!expandedKeys.value.includes(id)) {
        expandedKeys.value = [...expandedKeys.value, id]
      }
    } else {
      expandedKeys.value = expandedKeys.value.filter((key) => key !== id)
      allExpanded.value = false
    }
  }

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

  function reset() {
    events.value = []
    specs.value = {}
    isFinished.value = false
    finalStatus.value = 'pass'
    projectName.value = ''
    executionTime.value = 0
    environment.value = ''
    specsSkipped.value = 0
    expandedKeys.value = []
    nextId = 1
  }

  async function fetchInfo() {
    try {
      const res = await fetch('/api/info')
      if (res.ok) {
        const info = await res.json()
        timestamp.value = info.timestamp || timestamp.value
        if (info.project && !projectName.value) {
          projectName.value = info.project
        }
      }
    } catch {
      /* ignore */
    }
    if (!timestamp.value) {
      timestamp.value = formatTimestamp()
    }
  }

  function wsUrl() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/api/events/stream`
  }

  function connect() {
    closed = false
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return
    }
    if (isFinished.value) return
    socket = new WebSocket(wsUrl())

    socket.onopen = () => {
      isConnected.value = true
      reset()
    }

    socket.onmessage = (e) => {
      let data
      try {
        data = JSON.parse(e.data)
      } catch {
        return
      }
      addEvent(data)
    }

    socket.onclose = () => {
      isConnected.value = false
      if (!closed && !isFinished.value) {
        reconnectTimer = setTimeout(connect, 1000)
      }
    }

    socket.onerror = () => {
      socket?.close()
    }
  }

  function disconnect() {
    closed = true
    clearTimeout(reconnectTimer)
    socket?.close()
    socket = null
    isConnected.value = false
  }

  return {
    events,
    specs,
    specList,
    treeData,
    stats,
    isFinished,
    finalStatus,
    projectName,
    executionTime,
    environment,
    specsSkipped,
    timestamp,
    isConnected,
    allExpanded,
    expandedKeys,
    durationText,
    projectLabel,
    failed,
    passed,
    addEvent,
    reset,
    toggleAll,
    onExpandChange,
    fetchInfo,
    connect,
    disconnect
  }
})
