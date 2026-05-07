export interface ResourceUsage {
  categoryId: string;
  name: string;
  used: number;
  limit: number;
  unit: string;
  resetDate?: string;
}

export interface UsageData {
  platform: string;
  accountName: string;
  plan: string;
  resources: ResourceUsage[];
  lastUpdated: string;
  status: ServiceStatus;
  error?: string;
}

export type ServiceStatus = "ok" | "warn" | "danger" | "error";

export interface AlertInfo {
  platform: string;
  resource: string;
  percentage: number;
  time: string;
  status: "warn" | "danger" | "critical";
}

export interface DashboardData {
  services: UsageData[];
  alerts: AlertInfo[];
  lastSync: string;
  connectedCount: number;
  healthyCount: number;
  warnCount: number;
}

export interface AuthField {
  id: string;
  name: string;
  description: string;
  type: "text" | "password";
  required: boolean;
}

export interface AuthConfig {
  fields: AuthField[];
}

export interface AppSettings {
  theme: "light" | "dark" | "auto";
  syncInterval: number;
  alertThreshold: number;
}

export interface PlatformCredential {
  platform: string;
  accountName: string;
}

export interface ProjectInfo {
  id: string;
  name: string;
  url?: string;
  createdAt?: string;
}

export interface SyncResult {
  platform: string;
  success: boolean;
  error?: string;
}
