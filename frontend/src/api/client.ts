// Thin fetch wrapper shared by the API modules (see project instructions,
// section 44): components call an api/* function, never fetch() directly.
const API_BASE_URL = "http://localhost:8080";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, init);

  if (!response.ok) {
    throw new Error(await extractError(response));
  }

  return response.json() as Promise<T>;
}

async function extractError(response: Response): Promise<string> {
  try {
    const data: unknown = await response.json();
    if (data && typeof data === "object" && "error" in data && typeof data.error === "string") {
      return data.error;
    }
  } catch {
    // Body wasn't JSON; fall back to the status line below.
  }
  return `${response.status} ${response.statusText}`;
}

export function apiGet<T>(path: string): Promise<T> {
  return request<T>(path);
}

export function apiPost<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
}
