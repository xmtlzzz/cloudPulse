import React from "react";
import { ServiceCard } from "../components/ServiceCard";
import { UsageData, AlertInfo, ProjectInfo } from "../types";

interface DashboardProps {
  services: UsageData[];
  alerts: AlertInfo[];
  projects: Record<string, ProjectInfo[]>;
  lastSync: string;
  onServiceClick: (platform: string) => void;
}

function formatTimeAgo(dateStr: string): string {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffMin = Math.floor(diffMs / 60000);
  if (diffMin < 1) return "刚刚";
  if (diffMin < 60) return `${diffMin} 分钟前`;
  const diffHr = Math.floor(diffMin / 60);
  if (diffHr < 24) return `${diffHr} 小时前`;
  return `${Math.floor(diffHr / 24)} 天前`;
}

function getAlertStatusClass(status: string): string {
  switch (status) {
    case "critical":
      return "pill-danger";
    case "danger":
      return "pill-danger";
    case "warn":
      return "pill-warn";
    default:
      return "pill-warn";
  }
}

function getAlertStatusLabel(status: string): string {
  switch (status) {
    case "critical":
      return "超额";
    case "danger":
      return "临界";
    case "warn":
      return "注意";
    default:
      return "注意";
  }
}

export const Dashboard: React.FC<DashboardProps> = ({
  services,
  alerts,
  projects,
  lastSync,
  onServiceClick,
}) => {
  const healthyCount = services.filter((s) => s.status === "ok").length;
  const totalProjects = Object.values(projects).reduce((sum, p) => sum + p.length, 0);

  return (
    <div>
      {/* Summary stats */}
      <div className="grid-4" style={{ marginBottom: 24 }}>
        <div className="card">
          <div className="stat-label">已连接服务</div>
          <div className="stat-value">{services.length}</div>
        </div>
        <div className="card">
          <div className="stat-label">额度健康</div>
          <div className="stat-value" style={{ color: "var(--success)" }}>
            {healthyCount} / {services.length}
          </div>
        </div>
        <div className="card">
          <div className="stat-label">项目总数</div>
          <div className="stat-value">{totalProjects}</div>
        </div>
        <div className="card">
          <div className="stat-label">上次同步</div>
          <div className="stat-value" style={{ fontSize: 18, fontWeight: 500 }}>
            {formatTimeAgo(lastSync)}
          </div>
        </div>
      </div>

      {/* Service cards */}
      <div className="grid-2">
        {services.map((svc) => (
          <ServiceCard
            key={svc.platform}
            service={svc}
            projects={projects[svc.platform] || []}
            onClick={() => onServiceClick(svc.platform)}
          />
        ))}
      </div>

      {/* Recent alerts */}
      {alerts.length > 0 && (
        <div className="card" style={{ marginTop: 24 }}>
          <div className="card-header">
            <span className="card-title">近期提醒</span>
            <span className="card-subtitle">额度使用超过 80% 时触发</span>
          </div>
          <table className="table">
            <thead>
              <tr>
                <th>服务</th>
                <th>资源</th>
                <th>用量</th>
                <th>时间</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {alerts.map((alert, i) => (
                <tr key={i}>
                  <td>
                    <span style={{ fontWeight: 500 }}>{alert.platform}</span>
                  </td>
                  <td>{alert.resource}</td>
                  <td className="mono">{alert.percentage.toFixed(1)}%</td>
                  <td style={{ color: "var(--muted)" }}>
                    {formatTimeAgo(alert.time)}
                  </td>
                  <td>
                    <span className={`pill ${getAlertStatusClass(alert.status)}`}>
                      <span className="pill-dot" />
                      {getAlertStatusLabel(alert.status)}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};
