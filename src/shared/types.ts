export interface DevOptions {
  port?: number
  root?: string
  file?: string
  cors?: boolean
  https?: boolean
  historyApiFallback?: boolean

  proxy?: ProxyMiddleware[]

  server?: Server
}

export interface ProxyMiddleware {
  prefix: string
  target: string
  changeOrigin?: boolean
  secure?: boolean
}

export interface Server {
  cert: string
  key: string
}

export declare function dev(options: DevOptions): Promise<void>
