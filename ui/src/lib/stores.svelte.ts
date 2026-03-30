import { api } from './api';
import type { Config, PromptConfig, Job } from './types';

let config = $state<Config | null>(null);
let promptConfig = $state<PromptConfig | null>(null);
let jobs = $state<Job[]>([]);

export function getConfig() { return config; }
export function getPromptConfig() { return promptConfig; }
export function getJobs() { return jobs; }

export async function loadConfig() {
  config = await api.getConfig();
}

export async function loadPromptConfig() {
  promptConfig = await api.getPromptConfig();
}

export async function loadJobs() {
  jobs = await api.getJobs();
}

export async function refreshJobs() {
  jobs = await api.getJobs();
}
