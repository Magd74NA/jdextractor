export interface Config {
  deepseek_api_key: string;
  deepseek_model: string;
  kimi_api_key: string;
  kimi_model: string;
  backend: string;
  port: number;
}

export interface PromptConfig {
  system_prompt: string;
  task_list: string;
}

export interface Templates {
  resume: string;
  cover: string;
}

export interface Job {
  dir: string;
  company: string;
  role: string;
  score: number;
  status: string;
  tokens: number;
  date: string;
}

export interface JobFiles {
  resume: string;
  cover?: string;
}

export interface BatchResult {
  url: string;
  dir?: string;
  error?: string;
}

export interface ProcessResult {
  dir: string;
}

export interface ProgressEvent {
  stage: string;
  message?: string;
  dir?: string;
}

export type JobStatus = 'draft' | 'applied' | 'interviewing' | 'offer' | 'rejected';

export const JOB_STATUSES: JobStatus[] = ['draft', 'applied', 'interviewing', 'offer', 'rejected'];
