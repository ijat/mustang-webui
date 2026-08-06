import type { ApiError, InspectResponse } from './types';

export class ApiRequestError extends Error {}

/**
 * POSTs the raw PDF bytes to /api/inspect. Not multipart — the body is
 * the file itself, matching the sidecar contract exactly.
 */
export async function inspectPdf(file: File): Promise<InspectResponse> {
  const res = await fetch('/api/inspect', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/pdf',
      'X-Filename': encodeURIComponent(file.name),
    },
    body: file,
  });

  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = (await res.json()) as ApiError;
      if (body?.error) message = body.error;
    } catch {
      // Response body wasn't JSON — fall back to the generic message.
    }
    throw new ApiRequestError(message);
  }

  return (await res.json()) as InspectResponse;
}
