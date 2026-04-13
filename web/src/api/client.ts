export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path)

  if (!response.ok) {
    throw new Error(await getErrorMessage(response))
  }

  return response.json() as Promise<T>
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: body ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    throw new Error(await getErrorMessage(response))
  }

  return response.json() as Promise<T>
}

async function getErrorMessage(response: Response) {
  const text = (await response.text()).trim()
  if (text) {
    return text
  }

  return `Request failed: ${response.status}`
}
