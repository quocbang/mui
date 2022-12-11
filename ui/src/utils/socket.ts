let Socket: any
let setIntervalWesocketPush: any
const url = window.config.SocketUrl
let getSocketData: any

export const createSocket = () => {
  Socket && Socket.close()
  if (!Socket) {
    console.log('establish a websocket connection')
    Socket = new WebSocket(url)
    Socket.onmessage = onmessageWS
    Socket.onerror = onerrorWS
    Socket.onclose = oncloseWS
  } else {
    console.log('websocket connected')
  }
}

const onerrorWS = () => {
  Socket.close()
  clearInterval(setIntervalWesocketPush)
  console.log('connection failed reconnecting')
  if (Socket.readyState !== 3) {
    Socket = null
    createSocket()
  }
}

const onmessageWS = (e: { data: any }) => {
  window.dispatchEvent(new CustomEvent('onmessageWS', {
    detail: {
      data: e.data
    }
  }))
}

const connecting = (message: any) => {
  setTimeout(() => {
    if (Socket.readyState === 0) {
      connecting(message)
    } else {
      Socket.send(JSON.stringify(message))
    }
  }, 1000)
}

export const sendWSPush = (message: any) => {
  if (Socket !== null && Socket.readyState === 3) {
    Socket.close()
    createSocket()
  } else if (Socket.readyState === 1) {
    Socket.send(JSON.stringify(message))
  } else if (Socket.readyState === 0) {
    connecting(message)
  }
}

const oncloseWS = () => {
  clearInterval(setIntervalWesocketPush)
  console.log('websocket disconnected.... trying to reconnect')
  if (Socket.readyState !== 2) {
    Socket = null
    createSocket()
  }
}

export const webSocketCreate = () => {
  createSocket()
  getSocketData = (e: { detail: { data: any } }) => {
    const data = e && e.detail.data
    console.log(data)
  }
  window.addEventListener('onmessageWS', getSocketData)
}

export const WebSocketRemove = () => {
  window.removeEventListener('onmessageWS', getSocketData)
}
