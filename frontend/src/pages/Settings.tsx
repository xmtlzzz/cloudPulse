import React, { useState } from "react";
import { PlatformCredential, AppSettings, AuthConfig } from "../types";
import { TestConnection } from "../../wailsjs/go/main/App";

interface SettingsProps {
  credentials: PlatformCredential[];
  settings: AppSettings;
  availablePlatforms: { name: string; icon: string; description: string }[];
  onAddService: (platform: string) => void;
  onAddCustomService: (config: any) => void;
  onRemoveCredential: (platform: string, accountName: string) => void;
  onUpdateSettings: (settings: AppSettings) => void;
}

const iconMap: Record<string, { class: string; label: string }> = {
  Vercel: { class: "vercel", label: "V" },
  Neon: { class: "neon", label: "N" },
  Supabase: { class: "supabase", label: "S" },
  Cloudflare: { class: "cloudflare", label: "CF" },
};

export const Settings: React.FC<SettingsProps> = ({
  credentials,
  settings,
  availablePlatforms,
  onAddService,
  onAddCustomService,
  onRemoveCredential,
  onUpdateSettings,
}) => {
  const [localSettings, setLocalSettings] = useState(settings);
  const [showPlatformSelect, setShowPlatformSelect] = useState(false);
  const [showCustomServiceModal, setShowCustomServiceModal] = useState(false);
  const [testingPlatform, setTestingPlatform] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{ platform: string; success: boolean; message: string } | null>(null);

  const handleThemeChange = (theme: "light" | "dark" | "auto") => {
    const updated = { ...localSettings, theme };
    setLocalSettings(updated);
    onUpdateSettings(updated);
  };

  const handleSyncIntervalChange = (interval: number) => {
    const updated = { ...localSettings, syncInterval: interval };
    setLocalSettings(updated);
    onUpdateSettings(updated);
  };

  const handleAlertThresholdChange = (threshold: number) => {
    const updated = { ...localSettings, alertThreshold: threshold };
    setLocalSettings(updated);
    onUpdateSettings(updated);
  };

  const handlePlatformSelect = (platform: string) => {
    setShowPlatformSelect(false);
    onAddService(platform);
  };

  const handleTestConnection = async (platform: string) => {
    setTestingPlatform(platform);
    setTestResult(null);
    try {
      const result = await TestConnection(platform);
      setTestResult({
        platform,
        success: result.success,
        message: result.success
          ? `连接成功！账户: ${result.account}, 计划: ${result.plan}, 资源数: ${result.resources}`
          : `连接失败: ${result.error}`,
      });
    } catch (e: any) {
      setTestResult({
        platform,
        success: false,
        message: `测试失败: ${e.message}`,
      });
    }
    setTestingPlatform(null);
  };

  return (
    <div>
      {/* Token management */}
      <div className="settings-section">
        <div className="settings-section-title">Token 管理</div>
        <p style={{ fontSize: 13, color: "var(--muted)", marginBottom: 16 }}>
          粘贴各服务的 API Token 以启用用量监控。Token 仅存储在本地，不会上传至任何服务器。
        </p>

        {credentials.map((cred) => {
          const icon = iconMap[cred.platform] || { class: "", label: "?" };
          const isTesting = testingPlatform === cred.platform;
          const result = testResult?.platform === cred.platform ? testResult : null;
          return (
            <div key={`${cred.platform}-${cred.accountName}`}>
              <div className="token-row">
                <div className="token-svc">
                  <div className={`svc-icon ${icon.class}`} style={{ width: 28, height: 28, fontSize: 11 }}>
                    {icon.label}
                  </div>
                  <span style={{ fontWeight: 500, fontSize: 13 }}>{cred.platform}</span>
                </div>
                <span className="token-value">••••••••••••••••</span>
                <button
                  className="btn btn-sm"
                  onClick={() => handleTestConnection(cred.platform)}
                  disabled={isTesting}
                >
                  {isTesting ? "测试中..." : "测试连接"}
                </button>
                <button className="btn btn-sm">更新</button>
                <button
                  className="btn btn-sm btn-danger"
                  onClick={() => onRemoveCredential(cred.platform, cred.accountName)}
                >
                  移除
                </button>
              </div>
              {result && (
                <div
                  style={{
                    padding: "8px 12px",
                    fontSize: 12,
                    color: result.success ? "var(--success)" : "var(--danger)",
                    background: result.success ? "var(--success-subtle)" : "var(--danger-subtle)",
                    borderRadius: 4,
                    marginTop: 4,
                    marginBottom: 8,
                  }}
                >
                  {result.message}
                </div>
              )}
            </div>
          );
        })}

        <div style={{ marginTop: 16, display: "flex", gap: 8 }}>
          <button className="btn btn-primary" onClick={() => setShowPlatformSelect(true)}>
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M8 3v10M3 8h10" />
            </svg>
            添加新服务
          </button>
          <button className="btn" onClick={() => setShowCustomServiceModal(true)}>
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M2 2h12v12H2z" />
              <path d="M6 6h4M8 4v4" />
            </svg>
            自定义服务
          </button>
        </div>
      </div>

      {/* Platform selection modal */}
      {showPlatformSelect && (
        <div
          style={{
            position: "fixed",
            inset: 0,
            background: "rgba(0,0,0,0.5)",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            zIndex: 100,
          }}
        >
          <div className="card" style={{ width: 500, maxHeight: "70vh", overflow: "auto" }}>
            <div className="card-header">
              <span className="card-title">选择要添加的服务</span>
              <button
                className="btn btn-sm"
                onClick={() => setShowPlatformSelect(false)}
              >
                关闭
              </button>
            </div>
            <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
              {availablePlatforms.map((platform) => {
                const icon = iconMap[platform.name] || { class: "", label: platform.name.charAt(0) };
                const alreadyAdded = credentials.some((c) => c.platform === platform.name);
                return (
                  <button
                    key={platform.name}
                    className="btn"
                    style={{
                      justifyContent: "flex-start",
                      padding: "12px 16px",
                      opacity: alreadyAdded ? 0.5 : 1,
                      cursor: alreadyAdded ? "not-allowed" : "pointer",
                    }}
                    onClick={() => !alreadyAdded && handlePlatformSelect(platform.name)}
                    disabled={alreadyAdded}
                  >
                    <div className={`svc-icon ${icon.class}`} style={{ width: 32, height: 32, fontSize: 12 }}>
                      {icon.label}
                    </div>
                    <div style={{ textAlign: "left" }}>
                      <div style={{ fontWeight: 500 }}>{platform.name}</div>
                      <div style={{ fontSize: 12, color: "var(--muted)" }}>{platform.description}</div>
                    </div>
                    {alreadyAdded && (
                      <span className="pill pill-success" style={{ marginLeft: "auto", fontSize: 11 }}>
                        已添加
                      </span>
                    )}
                  </button>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Custom service modal */}
      {showCustomServiceModal && (
        <CustomServiceModal
          onSubmit={(config) => {
            onAddCustomService(config);
            setShowCustomServiceModal(false);
          }}
          onCancel={() => setShowCustomServiceModal(false)}
        />
      )}

      {/* Appearance */}
      <div className="settings-section">
        <div className="settings-section-title">外观</div>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 0" }}>
          <div>
            <div style={{ fontSize: 13, fontWeight: 500 }}>主题模式</div>
            <div style={{ fontSize: 12, color: "var(--muted)" }}>跟随系统或手动切换</div>
          </div>
          <div style={{ display: "flex", gap: 6 }}>
            <button
              className={`btn btn-sm ${localSettings.theme === "light" ? "btn-primary" : ""}`}
              onClick={() => handleThemeChange("light")}
            >
              浅色
            </button>
            <button
              className={`btn btn-sm ${localSettings.theme === "dark" ? "btn-primary" : ""}`}
              onClick={() => handleThemeChange("dark")}
            >
              深色
            </button>
            <button
              className={`btn btn-sm ${localSettings.theme === "auto" ? "btn-primary" : ""}`}
              onClick={() => handleThemeChange("auto")}
            >
              跟随系统
            </button>
          </div>
        </div>
      </div>

      {/* Sync settings */}
      <div className="settings-section">
        <div className="settings-section-title">同步设置</div>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 0" }}>
          <div>
            <div style={{ fontSize: 13, fontWeight: 500 }}>自动同步间隔</div>
            <div style={{ fontSize: 12, color: "var(--muted)" }}>定期从各服务拉取最新用量</div>
          </div>
          <select
            className="input"
            style={{ width: 140 }}
            value={localSettings.syncInterval}
            onChange={(e) => handleSyncIntervalChange(Number(e.target.value))}
          >
            <option value={5}>每 5 分钟</option>
            <option value={15}>每 15 分钟</option>
            <option value={30}>每 30 分钟</option>
            <option value={60}>每小时</option>
            <option value={0}>手动</option>
          </select>
        </div>
        <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "12px 0" }}>
          <div>
            <div style={{ fontSize: 13, fontWeight: 500 }}>额度预警阈值</div>
            <div style={{ fontSize: 12, color: "var(--muted)" }}>超过此百分比时在仪表盘高亮</div>
          </div>
          <select
            className="input"
            style={{ width: 100 }}
            value={localSettings.alertThreshold}
            onChange={(e) => handleAlertThresholdChange(Number(e.target.value))}
          >
            <option value={50}>50%</option>
            <option value={60}>60%</option>
            <option value={80}>80%</option>
            <option value={90}>90%</option>
          </select>
        </div>
      </div>

      {/* About */}
      <div className="settings-section">
        <div className="settings-section-title">关于</div>
        <div style={{ fontSize: 13, color: "var(--muted)" }}>
          CloudPulse v0.1.0 · 基于 Wails (Go) 构建 · 本地运行，数据不离开设备
        </div>
      </div>
    </div>
  );
};

interface CustomServiceModalProps {
  onSubmit: (config: any) => void;
  onCancel: () => void;
}

const CustomServiceModal: React.FC<CustomServiceModalProps> = ({ onSubmit, onCancel }) => {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [baseUrl, setBaseUrl] = useState("");
  const [authType, setAuthType] = useState("bearer");
  const [authKey, setAuthKey] = useState("");
  const [resources, setResources] = useState([
    { name: "", path: "", unit: "", defaultLimit: 0 },
  ]);

  const addResource = () => {
    setResources([...resources, { name: "", path: "", unit: "", defaultLimit: 0 }]);
  };

  const updateResource = (index: number, field: string, value: any) => {
    const updated = [...resources];
    (updated[index] as any)[field] = value;
    setResources(updated);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!name || !baseUrl) return;

    onSubmit({
      name,
      description,
      baseUrl,
      authType,
      authKey: authType === "bearer" ? "" : authKey,
      resources: resources.filter((r) => r.name && r.path),
    });
  };

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 100,
      }}
    >
      <div className="card" style={{ width: 600, maxHeight: "80vh", overflow: "auto" }}>
        <div className="card-header">
          <span className="card-title">添加自定义服务</span>
          <button className="btn btn-sm" onClick={onCancel}>
            关闭
          </button>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <label className="input-label">服务名称 *</label>
            <input
              className="input"
              type="text"
              placeholder="例如: My Custom API"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>

          <div className="input-group">
            <label className="input-label">描述</label>
            <input
              className="input"
              type="text"
              placeholder="简短描述这个服务"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>

          <div className="input-group">
            <label className="input-label">API 端点 *</label>
            <input
              className="input input-mono"
              type="text"
              placeholder="https://api.example.com/usage"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
              required
            />
          </div>

          <div className="input-group">
            <label className="input-label">认证方式</label>
            <select
              className="input"
              value={authType}
              onChange={(e) => setAuthType(e.target.value)}
            >
              <option value="bearer">Bearer Token</option>
              <option value="header">自定义 Header</option>
              <option value="query">Query 参数</option>
            </select>
          </div>

          {authType !== "bearer" && (
            <div className="input-group">
              <label className="input-label">
                {authType === "header" ? "Header 名称" : "Query 参数名"}
              </label>
              <input
                className="input"
                type="text"
                placeholder={authType === "header" ? "X-API-Key" : "api_key"}
                value={authKey}
                onChange={(e) => setAuthKey(e.target.value)}
              />
            </div>
          )}

          <div className="settings-section-title" style={{ marginTop: 20 }}>
            资源映射
          </div>
          <p style={{ fontSize: 12, color: "var(--muted)", marginBottom: 12 }}>
            定义如何从 API 响应中提取资源使用量
          </p>

          {resources.map((res, i) => (
            <div
              key={i}
              style={{
                display: "grid",
                gridTemplateColumns: "1fr 1fr 80px 100px",
                gap: 8,
                marginBottom: 8,
              }}
            >
              <input
                className="input"
                placeholder="资源名称"
                value={res.name}
                onChange={(e) => updateResource(i, "name", e.target.value)}
              />
              <input
                className="input input-mono"
                placeholder="JSON 路径 (如 data.bandwidth.used)"
                value={res.path}
                onChange={(e) => updateResource(i, "path", e.target.value)}
              />
              <input
                className="input"
                placeholder="单位"
                value={res.unit}
                onChange={(e) => updateResource(i, "unit", e.target.value)}
              />
              <input
                className="input"
                type="number"
                placeholder="限额"
                value={res.defaultLimit || ""}
                onChange={(e) => updateResource(i, "defaultLimit", Number(e.target.value))}
              />
            </div>
          ))}

          <button type="button" className="btn btn-sm" onClick={addResource} style={{ marginBottom: 16 }}>
            + 添加资源
          </button>

          <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
            <button type="button" className="btn" onClick={onCancel}>
              取消
            </button>
            <button type="submit" className="btn btn-primary">
              添加服务
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
