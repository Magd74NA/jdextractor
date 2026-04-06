import type { Config, PromptConfig, Templates, Job, JobFiles, BatchResult, ProcessResult, ProgressEvent, Contact, ConversationEntry, FollowupResult, NetworkingPromptConfig } from './types';

const BASE = '/api';

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const opts: RequestInit = {
    method,
    headers: {} as Record<string, string>,
  };
  if (body !== undefined) {
    (opts.headers as Record<string, string>)['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(`${BASE}${path}`, opts);
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return null as T;
  return res.json() as Promise<T>;
}

export const api = {
  getConfig: () => request<Config>('GET', '/config'),
  saveConfig: (data: Partial<Config>) => request<null>('PATCH', '/config', data),
  getPromptConfig: () => request<PromptConfig>('GET', '/config/prompt'),
  savePromptConfig: (data: Partial<PromptConfig>) => request<null>('PATCH', '/config/prompt', data),
  getTemplates: () => request<Templates>('GET', '/templates'),
  saveTemplates: (data: Partial<Templates>) => request<null>('PATCH', '/templates', data),
  getJobs: () => request<Job[]>('GET', '/jobs'),
  updateJobStatus: (id: string, status: string) => request<null>('PATCH', `/jobs/${id}`, { status }),
  updateJobMeta: (id: string, data: { company?: string; role?: string; date?: string }) =>
    request<null>('PATCH', `/jobs/${id}`, data),
  deleteJob: (id: string) => request<null>('DELETE', `/jobs/${id}`),
  getJobFiles: (id: string) => request<JobFiles>('GET', `/jobs/${id}/files`),
  saveJobFiles: (id: string, data: Partial<JobFiles>) => request<null>('PATCH', `/jobs/${id}/files`, data),
  process: (url: string) => request<ProcessResult>('POST', '/process', { url }),
  processBatch: (urls: string[]) => request<BatchResult[]>('POST', '/process/batch', { urls }),
  processLocal: (content: string) => request<ProcessResult>('POST', '/process/local', { content }),
  processStream: (url: string, onProgress: (event: ProgressEvent) => void) =>
    consumeSSE(`${BASE}/process/stream`, { url }, onProgress),
  processLocalStream: (content: string, onProgress: (event: ProgressEvent) => void) =>
    consumeSSE(`${BASE}/process/local/stream`, { content }, onProgress),

  // Contacts
  getContacts: () => request<Contact[]>('GET', '/contacts'),
  createContact: (data: Partial<Contact>) => request<{ dir: string }>('POST', '/contacts', data),
  getContact: (id: string) => request<Contact>('GET', `/contacts/${id}`),
  updateContact: (id: string, data: Partial<Contact>) => request<null>('PATCH', `/contacts/${id}`, data),
  deleteContact: (id: string) => request<null>('DELETE', `/contacts/${id}`),
  addConversation: (id: string, entry: ConversationEntry) =>
    request<null>('POST', `/contacts/${id}/conversations`, entry),
  deleteConversation: (id: string, index: number) =>
    request<null>('DELETE', `/contacts/${id}/conversations/${index}`),
  generateFollowup: (id: string) => request<FollowupResult>('POST', `/contacts/${id}/followup`),
  generateFollowupStream: async (id: string, onProgress: (event: ProgressEvent) => void): Promise<FollowupResult> => {
    const final = await consumeSSERaw(`${BASE}/contacts/${id}/followup/stream`, {}, onProgress);
    return JSON.parse(final.message ?? '{}') as FollowupResult;
  },
  getOverdueFollowups: () => request<Contact[]>('GET', '/contacts/overdue'),
  getUpcomingFollowups: (days?: number) =>
    request<Contact[]>('GET', `/contacts/upcoming${days !== undefined ? `?days=${days}` : ''}`),

  // Networking prompt config
  getNetworkingPromptConfig: () => request<NetworkingPromptConfig>('GET', '/config/networking-prompt'),
  saveNetworkingPromptConfig: (data: Partial<NetworkingPromptConfig>) =>
    request<null>('PATCH', '/config/networking-prompt', data),
};

async function consumeSSE(
  url: string,
  body: unknown,
  onProgress: (event: ProgressEvent) => void,
): Promise<ProcessResult> {
  const finalEvent = await consumeSSERaw(url, body, onProgress);
  return { dir: finalEvent.dir! };
}

export async function consumeSSERaw(
  url: string,
  body: unknown,
  onProgress: (event: ProgressEvent) => void,
): Promise<ProgressEvent> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  const reader = res.body!.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let finalEvent: ProgressEvent | null = null;

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });

    const lines = buffer.split('\n');
    buffer = lines.pop()!;

    for (const line of lines) {
      if (!line.startsWith('data: ')) continue;
      const event: ProgressEvent = JSON.parse(line.slice(6));
      if (event.stage === 'error') {
        throw new Error(event.message || 'Processing failed');
      }
      if (event.stage === 'complete') {
        finalEvent = event;
      }
      onProgress(event);
    }
  }

  if (!finalEvent) throw new Error('Stream ended without completion');
  return finalEvent;
}
