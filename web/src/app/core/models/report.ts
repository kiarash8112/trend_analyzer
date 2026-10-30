// Mirrors the JSON written by cmd/analyze_trending (services.Result and friends).

export interface ReportSummary {
  name: string;
  date: string;
  since: 'daily' | 'weekly' | 'monthly';
  repo_count: number;
}

export interface Repo {
  owner: string;
  name: string;
  url: string;
  description: string;
  language: string;
  topics?: string[];
  homepage?: string;
  license?: string;
  stars: number;
  forks: number;
  stars_in_period: number;
  period: string;
  built_by?: string[];
  created_at?: string;
}

export interface TimelineEvent {
  date: string;
  event: string;
  source_url: string;
}

export interface Insight {
  repo: string;
  analyzed_at: string;
  why_popular: string[] | string; // a single string in reports written before it became a list
  problem_solved: string;
  target_users: string;
  launch_channel: string;
  launch_url: string;
  launch_date: string;
  how_it_spread: string;
  timeline: TimelineEvent[];
  confidence: 'high' | 'medium' | 'low';
  sources: string[];
}

export interface RepoResult {
  evidence: { repo: Repo; notes?: string[] };
  insight?: Insight;
  cached: boolean;
  error?: string;
}
