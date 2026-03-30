import type { Config, PromptConfig, Templates, Job, JobFiles, BatchResult, ProcessResult } from './types';

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
  deleteJob: (id: string) => request<null>('DELETE', `/jobs/${id}`),
  getJobFiles: (id: string) => request<JobFiles>('GET', `/jobs/${id}/files`),
  saveJobFiles: (id: string, data: Partial<JobFiles>) => request<null>('PATCH', `/jobs/${id}/files`, data),
  process: (url: string) => request<ProcessResult>('POST', '/process', { url }),
  processBatch: (urls: string[]) => request<BatchResult[]>('POST', '/process/batch', { urls }),
  processLocal: (content: string) => request<ProcessResult>('POST', '/process/local', { content }),
};
