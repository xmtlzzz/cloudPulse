import { useState, useEffect, useCallback } from "react";
import "./style.css";
import { Sidebar } from "./components/Sidebar";
import { Topbar } from "./components/Topbar";
import { AddServiceModal } from "./components/AddServiceModal";
import { SyncResultModal } from "./components/SyncResultModal";
import { Dashboard } from "./pages/Dashboard";
import { ServiceDetail } from "./pages/ServiceDetail";
import { Settings } from "./pages/Settings";
import {
  UsageData,
  AlertInfo,
  PlatformCredential,
  AppSettings,
  AuthConfig,
  ServiceStatus,
  ProjectInfo,
  SyncResult,
} from "./types";
import {
  FetchAllUsage,
  FetchAllProjects,
  GetSettings,
  UpdateSettings,
  GetAllCredentials,
  AddCredential,
  RemoveCredential,
  GetAuthConfig,
  GetAvailablePlatforms,
  AddCustomService,
} from "../wailsjs/go/main/App";

const screenTitles: Record<string, string> = {
  dashboard: "仪表盘",
  "detail-vercel": "Vercel",
  "detail-neon": "Neon",
  "detail-supabase": "Supabase",
  "detail-cloudflare": "Cloudflare",
  settings: "设置",
};

// Mock data for initial display when no credentials are configured
const mockServices: UsageData[] = [
  {
    platform: "Vercel",
    accountName: "Personal",
    plan: "Hobby Plan",
    lastUpdated: new Date().toISOString(),
    status: "ok",
    resources: [
      { categoryId: "bandwidth", name: "带宽", used: 42.8, limit: 100, unit: "GB" },
      { categoryId: "builds", name: "构建次数", used: 186, limit: 6000, unit: "次" },
      { categoryId: "serverless", name: "Serverless 执行", used: 89400, limit: 100000, unit: "次" },
      { categoryId: "edge_functions", name: "Edge Function Invocations", used: 1204832, limit: 1000000, unit: "次" },
      { categoryId: "image_optimization", name: "Image Optimization", used: 2100, limit: 5000, unit: "次" },
    ],
  },
  {
    platform: "Neon",
    accountName: "Personal",
    plan: "Free Plan",
    lastUpdated: new Date().toISOString(),
    status: "ok",
    resources: [
      { categoryId: "storage", name: "存储", used: 287, limit: 512, unit: "MB" },
      { categoryId: "compute_time", name: "计算时间", used: 3.2, limit: 191.9, unit: "h" },
      { categoryId: "branches", name: "分支数", used: 3, limit: 10, unit: "个" },
    ],
  },
  {
    platform: "Supabase",
    accountName: "Personal",
    plan: "Free Plan",
    lastUpdated: new Date().toISOString(),
    status: "warn",
    resources: [
      { categoryId: "database_size", name: "数据库大小", used: 438, limit: 500, unit: "MB" },
      { categoryId: "bandwidth", name: "带宽", used: 1.8, limit: 5, unit: "GB" },
      { categoryId: "storage", name: "文件存储", used: 820, limit: 1024, unit: "MB" },
      { categoryId: "edge_functions", name: "Edge Function Invocations", used: 480000, limit: 500000, unit: "次" },
      { categoryId: "realtime_connections", name: "Realtime 并发连接", used: 180, limit: 200, unit: "个" },
    ],
  },
  {
    platform: "Cloudflare",
    accountName: "Personal",
    plan: "Free Plan",
    lastUpdated: new Date().toISOString(),
    status: "ok",
    resources: [
      { categoryId: "workers_requests", name: "Workers 请求", used: 68200, limit: 100000, unit: "次" },
      { categoryId: "r2_storage", name: "R2 存储", used: 4.1, limit: 10, unit: "GB" },
      { categoryId: "d1_database", name: "D1 数据库", used: 1.2, limit: 5, unit: "GB" },
      { categoryId: "pages_deployments", name: "Pages 部署", used: 124, limit: 500, unit: "次" },
      { categoryId: "kv_reads", name: "KV Reads", used: 82000, limit: 100000, unit: "次" },
      { categoryId: "kv_writes", name: "KV Writes", used: 1200, limit: 1000, unit: "次" },
    ],
  },
];

const mockAlerts: AlertInfo[] = [
  { platform: "Supabase", resource: "数据库大小", percentage: 87.6, time: new Date().toISOString(), status: "danger" },
  { platform: "Cloudflare", resource: "Workers 请求", percentage: 68.2, time: new Date(Date.now() - 3600000).toISOString(), status: "warn" },
  { platform: "Neon", resource: "存储用量", percentage: 56.1, time: new Date(Date.now() - 7200000).toISOString(), status: "warn" },
];

const mockCredentials: PlatformCredential[] = [
  { platform: "Vercel", accountName: "Personal" },
  { platform: "Neon", accountName: "Personal" },
  { platform: "Supabase", accountName: "Personal" },
  { platform: "Cloudflare", accountName: "Personal" },
];

function App() {
  const [activeScreen, setActiveScreen] = useState("dashboard");
  const [services, setServices] = useState<UsageData[]>(mockServices);
  const [alerts, setAlerts] = useState<AlertInfo[]>(mockAlerts);
  const [projects, setProjects] = useState<Record<string, ProjectInfo[]>>({
    Vercel: [
      { id: "1", name: "portfolio-site", url: "https://portfolio-site.vercel.app" },
      { id: "2", name: "saas-dashboard", url: "https://saas-dashboard.vercel.app" },
      { id: "3", name: "blog-engine", url: "https://blog-engine.vercel.app" },
      { id: "4", name: "marketing-v2", url: "https://marketing-v2.vercel.app" },
    ],
    Neon: [
      { id: "1", name: "my-saas-db" },
      { id: "2", name: "analytics-pool" },
    ],
    Supabase: [
      { id: "1", name: "my-saas", url: "https://my-saas.supabase.co" },
      { id: "2", name: "mobile-app-backend", url: "https://mobile-app-backend.supabase.co" },
    ],
    Cloudflare: [
      { id: "1", name: "api-gateway", url: "https://api-gateway.workers.dev" },
      { id: "2", name: "edge-cache", url: "https://edge-cache.workers.dev" },
      { id: "3", name: "static-site", url: "https://static-site.pages.dev" },
    ],
  });
  const [lastSync, setLastSync] = useState(new Date().toISOString());
  const [credentials, setCredentials] = useState<PlatformCredential[]>(mockCredentials);
  const [settings, setSettings] = useState<AppSettings>({
    theme: "auto",
    syncInterval: 15,
    alertThreshold: 80,
  });
  const [isDark, setIsDark] = useState(false);
  const [showAddModal, setShowAddModal] = useState(false);
  const [addPlatform, setAddPlatform] = useState("");
  const [addAuthConfig, setAddAuthConfig] = useState<AuthConfig | null>(null);
  const [showSyncResult, setShowSyncResult] = useState(false);
  const [syncResults, setSyncResults] = useState<SyncResult[]>([]);
  const [availablePlatforms, setAvailablePlatforms] = useState<{ name: string; icon: string; description: string }[]>([
    { name: "Vercel", icon: "vercel", description: "前端部署平台，提供 Serverless Functions 和 Edge Functions" },
    { name: "Neon", icon: "neon", description: "无服务器 Postgres 数据库平台" },
    { name: "Supabase", icon: "supabase", description: "开源 Firebase 替代方案，提供数据库、认证、存储等" },
    { name: "Cloudflare", icon: "cloudflare", description: "CDN、DNS 和边缘计算平台，提供 Workers、R2、D1 等" },
  ]);

  // Apply theme
  useEffect(() => {
    const applyTheme = () => {
      let resolved = settings.theme;
      if (resolved === "auto") {
        resolved = window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
      }
      document.documentElement.setAttribute("data-theme", resolved);
      setIsDark(resolved === "dark");
    };

    applyTheme();

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = () => {
      if (settings.theme === "auto") applyTheme();
    };
    mediaQuery.addEventListener("change", handler);
    return () => mediaQuery.removeEventListener("change", handler);
  }, [settings.theme]);

  // Load settings, credentials, and available platforms on mount
  useEffect(() => {
    const loadData = async () => {
      try {
        const savedSettings = await GetSettings();
        if (savedSettings) {
          setSettings({
            theme: (savedSettings.theme as "light" | "dark" | "auto") || "auto",
            syncInterval: savedSettings.syncInterval || 15,
            alertThreshold: savedSettings.alertThreshold || 80,
          });
        }
      } catch (e) {
        console.log("Using default settings");
      }

      try {
        const savedCreds = await GetAllCredentials();
        if (savedCreds && savedCreds.length > 0) {
          const creds: PlatformCredential[] = savedCreds.map((c: any) => ({
            platform: c.platform || "",
            accountName: c.accountName || "",
          }));
          setCredentials(creds);
          // Immediately fetch real data
          try {
            const data = await FetchAllUsage();
            if (data && data.services && data.services.length > 0) {
              const realServices = data.services.map((s: any) => ({
                ...s,
                status: (s.status as ServiceStatus) || "ok",
                lastUpdated: s.lastUpdated || new Date().toISOString(),
              }));
              const realAlerts = (data.alerts || []).map((a: any) => ({
                ...a,
                status: (a.status as "warn" | "danger" | "critical") || "warn",
                time: a.time || new Date().toISOString(),
              }));
              setServices(realServices as UsageData[]);
              setAlerts(realAlerts as AlertInfo[]);
              setLastSync(data.lastSync || new Date().toISOString());
            }
          } catch (fetchError) {
            console.log("Failed to fetch initial data:", fetchError);
          }

          // Fetch projects
          try {
            const projectsData = await FetchAllProjects();
            if (projectsData) {
              const projectsMap: Record<string, ProjectInfo[]> = {};
              for (const [platform, projectList] of Object.entries(projectsData)) {
                projectsMap[platform] = (projectList as any[]).map((p: any) => ({
                  id: p.id || "",
                  name: p.name || "",
                  url: p.url || "",
                  createdAt: p.createdAt || "",
                }));
              }
              setProjects(projectsMap);
            }
          } catch (projectError) {
            console.log("Failed to fetch projects:", projectError);
          }
        }
      } catch (e) {
        console.log("Using mock credentials");
      }

      try {
        const platforms = await GetAvailablePlatforms();
        if (platforms && platforms.length > 0) {
          setAvailablePlatforms(platforms.map((p: any) => ({
            name: p.name || "",
            icon: p.icon || "custom",
            description: p.description || "",
          })));
        }
      } catch (e) {
        console.log("Using default platforms");
      }
    };
    loadData();
  }, []);

  const refreshData = useCallback(async () => {
    try {
      const data = await FetchAllUsage();
      if (data) {
        const services = (data.services || []).map((s: any) => ({
          ...s,
          status: (s.status as ServiceStatus) || "ok",
          lastUpdated: s.lastUpdated || new Date().toISOString(),
        }));
        const alerts = (data.alerts || []).map((a: any) => ({
          ...a,
          status: (a.status as "warn" | "danger" | "critical") || "warn",
          time: a.time || new Date().toISOString(),
        }));
        setServices(services as UsageData[]);
        setAlerts(alerts as AlertInfo[]);
        setLastSync(data.lastSync || new Date().toISOString());
      }
    } catch (e) {
      console.log("Using mock data");
    }
  }, []);

  const handleSync = async () => {
    const results: SyncResult[] = [];

    // Actually fetch data from each service
    try {
      const data = await FetchAllUsage();
      if (data && data.services) {
        for (const service of data.services) {
          results.push({
            platform: service.platform,
            success: !service.error,
            error: service.error,
          });
        }
        // Update services data
        const updatedServices = data.services.map((s: any) => ({
          ...s,
          status: (s.status as ServiceStatus) || "ok",
          lastUpdated: s.lastUpdated || new Date().toISOString(),
        }));
        const updatedAlerts = (data.alerts || []).map((a: any) => ({
          ...a,
          status: (a.status as "warn" | "danger" | "critical") || "warn",
          time: a.time || new Date().toISOString(),
        }));
        setServices(updatedServices as UsageData[]);
        setAlerts(updatedAlerts as AlertInfo[]);
        setLastSync(data.lastSync || new Date().toISOString());
      }
    } catch (e: any) {
      // If FetchAllUsage fails, add error for all services
      for (const cred of credentials) {
        results.push({ platform: cred.platform, success: false, error: e.message });
      }
    }

    // Also fetch projects
    try {
      const projectsData = await FetchAllProjects();
      if (projectsData) {
        const projectsMap: Record<string, ProjectInfo[]> = {};
        for (const [platform, projectList] of Object.entries(projectsData)) {
          projectsMap[platform] = (projectList as any[]).map((p: any) => ({
            id: p.id || "",
            name: p.name || "",
            url: p.url || "",
            createdAt: p.createdAt || "",
          }));
        }
        setProjects(projectsMap);
      }
    } catch (e: any) {
      console.log("Failed to fetch projects:", e);
    }

    // Show results
    setSyncResults(results);
    setShowSyncResult(true);
  };

  const handleNavigate = (screen: string) => {
    setActiveScreen(screen);
  };

  const handleThemeToggle = () => {
    const newTheme = isDark ? "light" : "dark";
    setSettings((prev) => ({ ...prev, theme: newTheme }));
    UpdateSettings({ ...settings, theme: newTheme }).catch(() => {});
  };

  const handleServiceClick = (platform: string) => {
    setActiveScreen(`detail-${platform.toLowerCase()}`);
  };

  const handleAddService = async (platform: string) => {
    setAddPlatform(platform);

    // Get auth config for the selected platform
    try {
      const config = await GetAuthConfig(platform);
      if (config) {
        setAddAuthConfig({
          fields: (config.fields || []).map((f: any) => ({
            id: f.id || "",
            name: f.name || "",
            description: f.description || "",
            type: (f.type as "text" | "password") || "text",
            required: f.required || false,
          })),
        });
      }
    } catch (e) {
      // Fallback auth config
      setAddAuthConfig({
        fields: [
          { id: "token", name: "API Token", description: `${platform} API token`, type: "password", required: true },
        ],
      });
    }

    setShowAddModal(true);
  };

  const handleAddSubmit = async (accountName: string, creds: Record<string, string>) => {
    await AddCredential(addPlatform, accountName, creds);
    setCredentials((prev) => [...prev, { platform: addPlatform, accountName }]);
    setShowAddModal(false);
    refreshData();
  };

  const handleRemoveCredential = async (platform: string, accountName: string) => {
    try {
      await RemoveCredential(platform, accountName);
      setCredentials((prev) =>
        prev.filter((c) => !(c.platform === platform && c.accountName === accountName))
      );
    } catch (e) {
      console.error("Failed to remove credential:", e);
    }
  };

  const handleUpdateSettings = (newSettings: AppSettings) => {
    setSettings(newSettings);
    UpdateSettings(newSettings).catch(() => {});
  };

  const handleAddCustomService = async (config: any) => {
    try {
      await AddCustomService(config);
      // Refresh available platforms
      const platforms = await GetAvailablePlatforms();
      if (platforms && platforms.length > 0) {
        setAvailablePlatforms(platforms.map((p: any) => ({
          name: p.name || "",
          icon: p.icon || "custom",
          description: p.description || "",
        })));
      }
    } catch (e) {
      console.error("Failed to add custom service:", e);
    }
  };

  const getActiveService = (): UsageData | undefined => {
    const platformMap: Record<string, string> = {
      "detail-vercel": "Vercel",
      "detail-neon": "Neon",
      "detail-supabase": "Supabase",
      "detail-cloudflare": "Cloudflare",
    };
    const platform = platformMap[activeScreen];
    return services.find((s) => s.platform === platform);
  };

  const renderContent = () => {
    if (activeScreen === "dashboard") {
      return (
        <Dashboard
          services={services}
          alerts={alerts}
          projects={projects}
          lastSync={lastSync}
          onServiceClick={handleServiceClick}
        />
      );
    }

    if (activeScreen === "settings") {
      return (
        <Settings
          credentials={credentials}
          settings={settings}
          availablePlatforms={availablePlatforms}
          onAddService={handleAddService}
          onAddCustomService={handleAddCustomService}
          onRemoveCredential={handleRemoveCredential}
          onUpdateSettings={handleUpdateSettings}
        />
      );
    }

    if (activeScreen.startsWith("detail-")) {
      const service = getActiveService();
      const platformMap: Record<string, string> = {
        "detail-vercel": "Vercel",
        "detail-neon": "Neon",
        "detail-supabase": "Supabase",
        "detail-cloudflare": "Cloudflare",
      };
      const platform = platformMap[activeScreen];
      if (service) {
        return (
          <ServiceDetail
            service={service}
            projects={projects[platform] || []}
            onBack={() => handleNavigate("dashboard")}
          />
        );
      }
    }

    return null;
  };

  return (
    <div className="app">
      <Sidebar activeScreen={activeScreen} onNavigate={handleNavigate} />
      <div className="main">
        <Topbar
          title={screenTitles[activeScreen] || activeScreen}
          onSettingsClick={() => handleNavigate("settings")}
          onThemeToggle={handleThemeToggle}
          onSync={handleSync}
          isDark={isDark}
        />
        <div className="content">{renderContent()}</div>
      </div>

      {showAddModal && (
        <AddServiceModal
          platform={addPlatform}
          authConfig={addAuthConfig}
          onSubmit={handleAddSubmit}
          onCancel={() => setShowAddModal(false)}
        />
      )}

      {showSyncResult && (
        <SyncResultModal
          results={syncResults}
          onClose={() => setShowSyncResult(false)}
        />
      )}
    </div>
  );
}

export default App;
