export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(path, {
    credentials: "same-origin",
  });

  if (!response.ok) {
    throw await buildApiError(response);
  }

  return response.json() as Promise<T>;
}

export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!response.ok) {
    throw await buildApiError(response);
  }

  return response.json() as Promise<T>;
}

export async function apiUpload<T>(path: string, file: File): Promise<T> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(path, {
    method: "POST",
    credentials: "same-origin",
    body: formData,
  });

  if (!response.ok) {
    throw await buildApiError(response);
  }

  return response.json() as Promise<T>;
}

async function buildApiError(response: Response) {
  const message = await getErrorMessage(response);
  if (response.status === 401 && window.location.pathname !== "/login") {
    window.location.assign("/login");
  }

  return new ApiError(response.status, message);
}

async function getErrorMessage(response: Response) {
  const text = (await response.text()).trim();
  if (text) {
    return text;
  }

  return `Request failed: ${response.status}`;
}
