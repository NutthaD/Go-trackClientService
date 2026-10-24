import type { Client, ServiceUpdate } from './types'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const PAGE_SIZE = 100

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  })

  if (!response.ok) {
    const responseText = await response.text()
    let message = responseText
    try {
      const payload: unknown = JSON.parse(responseText)
      if (
        typeof payload === 'object' &&
        payload !== null &&
        'error' in payload &&
        typeof payload.error === 'string'
      ) {
        message = payload.error
      }
    } catch {
      message = responseText
    }
    if (response.status >= 500 && message === 'เกิดข้อผิดพลาดภายในระบบ') {
      message += ' — ตรวจรายละเอียด "update service failed" ใน terminal ของ Go API'
    }
    throw new Error(`${message || 'คำขอไม่สำเร็จ'} (HTTP ${response.status})`)
  }

  return response.json() as Promise<T>
}

export async function getClients(): Promise<Client[]> {
  const clients: Client[] = []
  let offset = 0

  while (true) {
    const result = await request<{ data: Client[] }>(
      `/clients?limit=${PAGE_SIZE}&offset=${offset}`,
    )
    clients.push(...result.data)
    if (result.data.length < PAGE_SIZE) return clients
    offset += result.data.length
  }
}

export function updateService(
  clientID: string,
  clientServiceID: string,
  update: ServiceUpdate,
): Promise<{ data: unknown }> {
  return request(
    `/clients/${encodeURIComponent(clientID)}/services/${encodeURIComponent(clientServiceID)}`,
    {
      method: 'PATCH',
      body: JSON.stringify(update),
    },
  )
}
