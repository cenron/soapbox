export class ApiError extends Error {
  readonly status: number
  readonly kind: string

  constructor(status: number, kind: string, message: string) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.kind = kind
  }

  get isUnauthorized(): boolean {
    return this.status === 401
  }

  get isForbidden(): boolean {
    return this.status === 403
  }

  get isNotFound(): boolean {
    return this.status === 404
  }
}
