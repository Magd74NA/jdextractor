import type { Job, Contact } from './types';

export interface DashboardStats {
  total: number;
  applied: number;
  avgScore: number;
  thisWeek: number;
}

export interface StatusCounts {
  draft: number;
  applied: number;
  interviewing: number;
  offer: number;
  rejected: number;
}

const DAY_S = 86400;

function parseDate(dateStr: string): number {
  const [y, m, d] = dateStr.split('-').map(Number);
  return Date.UTC(y!, m! - 1, d!) / 1000;
}

export function computeStats(jobs: Job[]): DashboardStats {
  const total = jobs.length;
  const applied = jobs.filter(j => j.status && j.status !== 'draft').length;
  const avgScore = total > 0
    ? Math.round((jobs.reduce((s, j) => s + j.score, 0) / total) * 10) / 10
    : 0;

  const weekAgo = Date.now() / 1000 - 7 * DAY_S;
  const thisWeek = jobs.filter(j => parseDate(j.date) >= weekAgo).length;

  return { total, applied, avgScore, thisWeek };
}

export function computeStatusCounts(jobs: Job[]): StatusCounts {
  const counts: StatusCounts = { draft: 0, applied: 0, interviewing: 0, offer: 0, rejected: 0 };
  for (const j of jobs) {
    const s = (j.status || 'draft') as keyof StatusCounts;
    if (s in counts) counts[s]++;
  }
  return counts;
}

export function computeActivityData(jobs: Job[]): [number[], number[]] {
  if (jobs.length === 0) return [[], []];

  const grouped = new Map<number, number>();
  for (const j of jobs) {
    const ts = parseDate(j.date);
    grouped.set(ts, (grouped.get(ts) || 0) + 1);
  }

  const timestamps = [...grouped.keys()].sort((a, b) => a - b);
  const min = timestamps[0]!;
  const max = timestamps[timestamps.length - 1]!;

  const xs: number[] = [];
  const ys: number[] = [];
  for (let t = min; t <= max; t += DAY_S) {
    xs.push(t);
    ys.push(grouped.get(t) || 0);
  }

  return [xs, ys];
}

export interface NetworkingStats {
  totalContacts: number;
  activeContacts: number;
  overdueFollowups: number;
  upcomingFollowups: number;
}

export function computeNetworkingStats(contacts: Contact[]): NetworkingStats {
  const today = new Date().toISOString().slice(0, 10);
  const weekFromNow = new Date(Date.now() + 7 * DAY_S * 1000).toISOString().slice(0, 10);
  return {
    totalContacts: contacts.length,
    activeContacts: contacts.filter(c => c.status !== 'dormant').length,
    overdueFollowups: contacts.filter(c => c.follow_up_date && c.follow_up_date <= today).length,
    upcomingFollowups: contacts.filter(
      c => c.follow_up_date && c.follow_up_date > today && c.follow_up_date <= weekFromNow,
    ).length,
  };
}

export function computeScoreDistribution(jobs: Job[]): [number[], number[]] {
  if (jobs.length === 0) return [[], []];

  const counts = new Array(10).fill(0) as number[];
  for (const j of jobs) {
    const idx = Math.max(0, Math.min(9, j.score - 1));
    counts[idx] = (counts[idx] ?? 0) + 1;
  }

  const xs = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
  return [xs, counts];
}
