import React from "react";
import { UsageData, ProjectInfo } from "../types";

interface ServiceDetailProps {
  service: UsageData;
  projects: ProjectInfo[];
  onBack: () => void;
}

const iconMap: Record<string, { class: string; label: string }> = {
  Vercel: { class: "vercel", label: "▲" },
  Neon: { class: "neon", label: "N" },
  Supabase: { class: "supabase", label: "S" },
  Cloudflare: { class: "cloudflare", label: "CF" },
};

function getStatusClass(percentage: number): string {
  if (percentage >= 90) return "danger";
  if (percentage >= 70) return "warn";
  return "ok";
}

function getStatusPill(used: number, limit: number): { class: string; label: string } {
  if (limit === 0) return { class: "pill-success", label: "正常" };
  const pct = (used / limit) * 100;
  if (pct >= 100) return { class: "pill-danger", label: "超额" };
  if (pct >= 90) return { class: "pill-danger", label: "临界" };
  if (pct >= 70) return { class: "pill-warn", label: "注意" };
  return { class: "pill-success", label: "正常" };
}

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(2) + "M";
  if (num >= 1000) return (num / 1000).toFixed(2) + "K";
  return num.toFixed(2);
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

export const ServiceDetail: React.FC<ServiceDetailProps> = ({ service, projects, onBack }) => {
  const icon = iconMap[service.platform] || { class: "", label: service.platform.charAt(0) };
  const mainResources = service.resources.slice(0, 3);
  const otherResources = service.resources.slice(3);

  return (
    <div>
      <button className="back-btn" onClick={onBack}>
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M10 3L5 8l5 5" />
        </svg>
        返回仪表盘
      </button>

      <div className="detail-header">
        <div className={`svc-icon ${icon.class}`}>{icon.label}</div>
        <div>
          <div className="detail-title">{service.platform}</div>
          <div className="detail-plan">
            {service.plan} · 免费额度 · 上次同步 {formatTimeAgo(service.lastUpdated)}
          </div>
        </div>
        <div style={{ marginLeft: "auto" }}>
          <span className={`pill ${service.status === "ok" ? "pill-success" : service.status === "warn" ? "pill-warn" : "pill-danger"}`}>
            <span className="pill-dot" />
            {service.status === "ok" ? "正常" : service.status === "warn" ? "注意" : "临界"}
          </span>
        </div>
      </div>

      {/* Main resource cards */}
      <div className="grid-3" style={{ marginBottom: 24 }}>
        {mainResources.map((res) => {
          const pct = res.limit > 0 ? (res.used / res.limit) * 100 : 0;
          const statusClass = getStatusClass(pct);
          return (
            <div className="card" key={res.categoryId}>
              <div className="stat-label">{res.name}</div>
              <div className="stat-value">
                {formatNumber(res.used)} {res.unit}
              </div>
              <div className="stat-label" style={{ marginTop: 4 }}>
                {res.limit > 0
                  ? `/ ${formatNumber(res.limit)} ${res.unit} · ${pct.toFixed(2)}%`
                  : "无限制"}
              </div>
              {res.limit > 0 && (
                <div className="progress-track">
                  <div
                    className={`progress-fill ${statusClass}`}
                    style={{ width: `${Math.min(pct, 100)}%` }}
                  />
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Detailed usage table */}
      {otherResources.length > 0 && (
        <div className="card">
          <div className="card-header">
            <span className="card-title">用量明细</span>
          </div>
          {otherResources.map((res) => {
            const pill = getStatusPill(res.used, res.limit);
            return (
              <div className="usage-row" key={res.categoryId}>
                <span className="usage-name">{res.name}</span>
                <span className="usage-detail">
                  {formatNumber(res.used)} / {res.limit > 0 ? formatNumber(res.limit) : "无限制"}
                </span>
                <span className={`pill ${pill.class}`}>
                  <span className="pill-dot" />
                  {pill.label}
                </span>
              </div>
            );
          })}
        </div>
      )}

      {/* Show all resources in table if no split needed */}
      {otherResources.length === 0 && mainResources.length > 0 && (
        <div className="card">
          <div className="card-header">
            <span className="card-title">用量明细</span>
          </div>
          {service.resources.map((res) => {
            const pill = getStatusPill(res.used, res.limit);
            return (
              <div className="usage-row" key={res.categoryId}>
                <span className="usage-name">{res.name}</span>
                <span className="usage-detail">
                  {formatNumber(res.used)} / {res.limit > 0 ? formatNumber(res.limit) : "无限制"}
                </span>
                <span className={`pill ${pill.class}`}>
                  <span className="pill-dot" />
                  {pill.label}
                </span>
              </div>
            );
          })}
        </div>
      )}

      {/* Projects */}
      {projects.length > 0 && (
        <div className="card" style={{ marginTop: 24 }}>
          <div className="card-header">
            <span className="card-title">项目 ({projects.length})</span>
            <span className="card-subtitle">该账号下的 {service.platform} 项目</span>
          </div>
          {projects.map((project) => (
            <div className="project-card" key={project.id}>
              <div className={`project-icon ${icon.class}`} style={{ background: "var(--border)", color: "var(--muted)" }}>
                {icon.label}
              </div>
              <div className="project-info">
                <div className="project-name">{project.name}</div>
                <div className="project-meta">
                  {project.createdAt ? `创建于 ${formatTimeAgo(project.createdAt)}` : "项目"}
                </div>
              </div>
              {project.url && (
                <a className="project-link" href={project.url} target="_blank" rel="noopener noreferrer">
                  {new URL(project.url).hostname}
                </a>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};
