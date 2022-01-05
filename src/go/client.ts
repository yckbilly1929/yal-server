import { createSnackbar } from '@snackbar/core'
import '@snackbar/core/dist/snackbar.css'

console.log('[yal-server] init')

if ('WebSocket' in window) {
  window.addEventListener('load', () => {
    const protocol = window.location.protocol === 'http:' ? 'ws:' : 'wss:'
    const host = 'localhost:{{port}}'
    const address = `${protocol}//${host}`

    const CONNECTING_MSG = '[yal-server] connecting...'
    const CONNECTED_MSG = '[yal-server] connected.'
    const MAX_ATTEMPTS = 30

    let wait = 1000
    let attempts = 0
    let socket!: WebSocket

    const refreshCSS = (showPopup: boolean) => {
      const head = document.getElementsByTagName('head')[0]

      let sheets = Array.from(document.getElementsByTagName('link'))
      sheets = sheets.filter(sheet => /\.css/gm.test(sheet.href) || sheet.rel.toLowerCase() == 'stylesheet')

      for (let i = 0; i < sheets.length; ++i) {
        const el = sheets[i]

        const newEl = el.cloneNode(true) as HTMLLinkElement

        // changing the href of the css file will make the browser refetch it
        const url = newEl.href.replace(/(&|\?)_cacheOverride=\d+/, '')
        newEl.href = `${url}${url.indexOf('?') >= 0 ? '&' : '?'}_cacheOverride=${new Date().valueOf()}`

        newEl.onload = () => {
          setTimeout(() => el.remove(), 0)
        }

        head.appendChild(newEl)
      }

      if (sheets.length > 0 && showPopup) {
        createSnackbar('css updated', {
          timeout: 5000,
        })
      }
    }

    const connect = () => {
      console.log(CONNECTING_MSG)
      socket = new WebSocket(address)

      socket.onmessage = function (msg) {
        // reset health check
        wait = 1000
        attempts = 0

        // console.log(`debug msg: type=${msg.type}; data=${msg.data}; origin: ${msg.origin}`)

        switch (msg.data) {
          case 'reload':
            window.location.reload()
            break
          case 'refresh-css':
            refreshCSS(false)
            break
          case 'refresh-css-popup':
            refreshCSS(true)
            break
          case 'connected':
            // TODO:
            console.log(CONNECTED_MSG)
            break
          case 'toggle-log':
            // TODO: toggle log between browser and shell
            break
          default:
            // TODO: toast message?
            break
        }
      }
      socket.onopen = function () {
        // reload page on successful reconnection
        if (attempts > 0) {
          window.location.reload()
          return
        }
      }
      socket.onclose = function (e) {
        if (attempts === 0) {
          console.log('[yal-server] socket closed: ', e.reason)
          createSnackbar('Lost connection to dev server...')
        }

        // TODO: exponential backoff
        setTimeout(function () {
          attempts++
          if (attempts <= MAX_ATTEMPTS) {
            connect()
          } else {
            console.log('[yal-server] reconnection stopped.')
            return
          }
          wait = Math.floor(wait * 1.1)
        }, wait)
      }
      socket.onerror = function (_e) {
        socket.close()
      }
    }

    connect()
  })
} else {
  // TODO
  console.error('WebSocket not supported')
}
