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
  delta?: string;
}

export type JobStatus = 'draft' | 'applied' | 'interviewing' | 'offer' | 'rejected';

export const JOB_STATUSES: JobStatus[] = ['draft', 'applied', 'interviewing', 'offer', 'rejected'];

export interface Message {
  sender: string;
  content: string;
  date: string;
  generated?: boolean;
}

export interface Conversation {
  channel?: string;
  summary: string;
  messages: Message[];
  created: string;
}

export interface Contact {
  dir: string;
  name: string;
  company?: string;
  role?: string;
  email?: string;
  phone?: string;
  linkedin?: string;
  source?: string;
  status: string;
  follow_up_date?: string;
  linked_jobs?: string[];
  tags?: string[];
  notes?: string;
  conversations: Conversation[];
  created: string;
}

export interface FollowupResult {
  subject?: string;
  message: string;
  channel: string;
  timing: string;
  notes: string;
  suggested_next_date?: string;
}

export interface NetworkingPromptConfig {
  system_prompt: string;
  task_list: string;
}

export type ContactStatus = 'new' | 'reached-out' | 'replied' | 'meeting-scheduled' | 'connected' | 'dormant';

export const CONTACT_STATUSES: ContactStatus[] = [
  'new', 'reached-out', 'replied', 'meeting-scheduled', 'connected', 'dormant',
];

export const CHANNELS = ['email', 'linkedin', 'phone', 'in-person', 'event', 'other'] as const;
