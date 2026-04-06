import { api } from './api';
import type { Config, PromptConfig, Job, Contact, NetworkingPromptConfig } from './types';

let config = $state<Config | null>(null);
let promptConfig = $state<PromptConfig | null>(null);
let jobs = $state<Job[]>([]);
let contacts = $state<Contact[]>([]);
let networkingPromptConfig = $state<NetworkingPromptConfig | null>(null);

export function getConfig() { return config; }
export function getPromptConfig() { return promptConfig; }
export function getJobs() { return jobs; }
export function getContacts() { return contacts; }
export function getNetworkingPromptConfig() { return networkingPromptConfig; }

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

export async function loadContacts() {
  contacts = await api.getContacts();
}

export async function refreshContacts() {
  contacts = await api.getContacts();
}

export async function loadNetworkingPromptConfig() {
  networkingPromptConfig = await api.getNetworkingPromptConfig();
}
