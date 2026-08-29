import { ref, onMounted, onUnmounted } from 'vue'

export function useWebSocket(url, { onEvent, onFinished, onOpen } = {}) {
  const isConnected = ref(false)
  const isFinished = ref(false)
  let socket = null
  let reconnectTimer = null
  let closed = false

  const connect = () => {
    if (closed || isFinished.value) return
    socket = new WebSocket(url)

    socket.onopen = () => {
      isConnected.value = true
      onOpen?.()
    }

    socket.onmessage = (e) => {
      let data
      try {
        data = JSON.parse(e.data)
      } catch {
        return
      }
      onEvent?.(data)
      if (data.type === 'end') {
        isFinished.value = true
        onFinished?.(data)
      }
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

  onMounted(connect)
  onUnmounted(() => {
    closed = true
    clearTimeout(reconnectTimer)
    socket?.close()
  })

  return { isConnected, isFinished }
}
