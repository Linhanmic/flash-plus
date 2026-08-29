import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(url) {
  const events = ref([])
  const isConnected = ref(false)
  const isFinished = ref(false)
  let socket = null

  const connect = () => {
    socket = new WebSocket(url)

    socket.onopen = () => {
      isConnected.value = true
      console.log('[Flash] WebSocket connected')
    }

    socket.onmessage = (e) => {
      const data = JSON.parse(e.data)
      events.value.push(data)

      if (data.type === 'end') {
        isFinished.value = true
      }
    }

    socket.onclose = () => {
      isConnected.value = false
      console.log('[Flash] WebSocket disconnected')
    }

    socket.onerror = (err) => {
      console.error('[Flash] WebSocket error:', err)
    }
  }

  onMounted(connect)
  onUnmounted(() => socket?.close())

  return { events, isConnected, isFinished }
}
