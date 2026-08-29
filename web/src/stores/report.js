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

function isConceptNode(item) {
  return item?.type === 'concept' || item?.concept === true
}

function lastProgressConcept(items) {
  if (!items?.length) return null
  for (let i = items.length - 1; i >= 0; i--) {
    const item = items[i]
    if (isConceptNode(item) && item.status === 'progress') {
      return lastProgressConcept(item.steps) || item
    }
  }
  return null
}

function findOpen(items, name, type) {
  if (!items?.length) return null
  for (let i = items.length - 1; i >= 0; i--) {
    const item = items[i]
    if (isConceptNode(item)) {
      const nested = findOpen(item.steps, name, type)
      if (nested) return nested
    }
    const itemType = isConceptNode(item) ? 'concept' : item.type || 'step'
    if (item.name === name && item.status === 'progress' && itemType === type) {
      return item
    }
  }
  return null
}

function targetList(scenario) {
  const concept = lastProgressConcept(scenario.steps)
  if (concept) {
    if (!concept.steps) concept.steps = []
    return concept.steps
  }
  return scenario.steps
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

function baseName(path) {
  if (!path) return ''
  return String(path).replace(/^.*[/\\]/, '')
}

function mapStep(item, level) {
  const concept = isConceptNode(item)
  const node = {
    id: `${concept ? 'cpt' : 'step'}-${item.id}`,
    name: item.name,
    kind: concept ? '概念' : '步骤',
    type: concept ? 'concept' : 'step',
    level,
    status: item.status || 'progress',
    errorMessage: item.errorMessage,
    stackTrace: item.stackTrace,
    children: []
  }
  if (concept) {
    for (const child of item.steps || []) {
      node.children.push(mapStep(child, level + 1))
    }
  }
  if (item.errorMessage || item.stackTrace) {
    node.children.push({
      id: `err-${item.id}`,
      name: item.errorMessage || 'Step failed',
      kind: '',
      type: 'error',
      level,
      status: 'fail',
      stackTrace: item.stackTrace
    })
  }
  if (!node.children.length) {
    delete node.children
  }
  return node
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
    specList.value.map((spec) => {
      const fileLabel = baseName(spec.fileName)
      const heading = spec.heading || spec.name
      return {
        id: `spec-${spec.id}`,
        name: fileLabel || heading || '(unknown spec)',
        heading: fileLabel && heading && heading !== fileLabel ? heading : '',
        fileName: spec.fileName,
        kind: '规格书',
        type: 'spec',
        level: 0,
        status: spec.status || 'progress',
        children: (spec.scenarios || []).map((scenario) => ({
          id: `scn-${scenario.id}`,
          name: scenario.name,
          kind: '场景',
          type: 'scenario',
          level: 1,
          status: scenario.status || 'progress',
          children: (scenario.steps || []).map((step) => mapStep(step, 2))
        }))
      }
    })
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

  function resolveScenario(spec, name) {
    return lastOpen(spec.scenarios, name) || spec.scenarios.find((s) => s.name === name) || null
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
        if (!specs.value[key]) {
          specs.value[key] = {
            id: uid(),
            name: event.name,
            heading: event.name,
            fileName: event.fileName,
            status: event.status,
            tags: event.tags,
            scenarios: []
          }
        } else {
          specs.value[key].status = event.status
          if (event.name) {
            specs.value[key].name = event.name
            specs.value[key].heading = event.name
          }
          if (event.fileName) specs.value[key].fileName = event.fileName
        }
        break
      }

      case 'scenario': {
        const spec = specs.value[event.specFileName]
        if (!spec) break
        if (event.status === 'progress') {
          spec.scenarios.push({
            id: uid(),
            name: event.name,
            status: event.status,
            steps: []
          })
        } else {
          const open = lastOpen(spec.scenarios, event.name)
          if (open) {
            open.status = event.status
          } else {
            spec.scenarios.push({ id: uid(), name: event.name, status: event.status, steps: [] })
          }
        }
        break
      }

      case 'concept': {
        const spec = specs.value[event.specFileName]
        if (!spec) break
        const scenario = resolveScenario(spec, event.scenarioName)
        if (!scenario) break
        if (event.status === 'progress') {
          targetList(scenario).push({
            id: uid(),
            type: 'concept',
            name: event.name,
            status: event.status,
            steps: []
          })
        } else {
          const open = findOpen(scenario.steps, event.name, 'concept')
          if (open) {
            open.status = event.status
            open.errorMessage = event.errorMessage
            open.stackTrace = event.stackTrace
          } else {
            scenario.steps.push({
              id: uid(),
              type: 'concept',
              name: event.name,
              status: event.status,
              steps: []
            })
          }
        }
        break
      }

      case 'step': {
        const spec = specs.value[event.specFileName]
        if (!spec) break
        const scenario = resolveScenario(spec, event.scenarioName)
        if (!scenario) break
        if (event.status === 'progress') {
          targetList(scenario).push({
            id: uid(),
            type: 'step',
            name: event.name,
            status: event.status
          })
        } else {
          const open = findOpen(scenario.steps, event.name, 'step')
          if (open) {
            open.status = event.status
            open.errorMessage = event.errorMessage
            open.stackTrace = event.stackTrace
          } else {
            targetList(scenario).push({
              id: uid(),
              type: 'step',
              name: event.name,
              status: event.status,
              errorMessage: event.errorMessage,
              stackTrace: event.stackTrace
            })
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
