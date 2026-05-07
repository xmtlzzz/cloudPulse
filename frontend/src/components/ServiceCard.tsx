import React from "react";
import { UsageData, ProjectInfo } from "../types";

interface ServiceCardProps {
  service: UsageData;
  projects: ProjectInfo[];
  onClick: () => void;
}

const iconMap: Record<string, { class: string; label: string }> = {
  Vercel: { class: "vercel", label: "V" },
  Neon: { class: "neon", label: "N" },
  Supabase: { class: "supabase", label: "S" },
  Cloudflare: { class: "cloudflare", label: "CF" },
};

function getStatusClass(percentage: number): string {
  if (percentage >= 90) return "danger";
  if (percentage >= 70) return "warn";
  return "ok";
}

function getStatusPill(status: string): { class: string; label: string } {
  switch (status) {
    case "ok":
      return { class: "pill-success", label: "正常" };
    case "warn":
      return { class: "pill-warn", label: "注意" };
    case "danger":
      return { class: "pill-danger", label: "临界" };
    case "error":
      return { class: "pill-danger", label: "错误" };
    default:
      return { class: "pill-success", label: "正常" };
  }
}

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + "M";
  if (num >= 1000) return (num / 1000).toFixed(1) + "K";
  return num.toString();
}

export const ServiceCard: React.FC<ServiceCardProps> = ({ service, projects, onClick }) => {
  const icon = iconMap[service.platform] || { class: "", label: service.platform.charAt(0) };
  const pill = getStatusPill(service.status);

  // Get top 2 resources to display
  const topResources = service.resources.filter((r) => r.limit > 0).slice(0, 2);

  return (
    <div className="svc-card" onClick={onClick}>
      <div className="svc-card-top">
        <div className={`svc-icon ${icon.class}`}>{icon.label}</div>
        <div>
          <div className="svc-card-name">{service.platform}</div>
          <div className="svc-card-plan">{service.plan} — 免费</div>
        </div>
        <div style={{ marginLeft: "auto" }}>
          <span className={`pill ${pill.class}`}>
            <span className="pill-dot" />
            {pill.label}
          </span>
        </div>
      </div>
      <div className="svc-card-stats">
        {topResources.map((res) => {
          const pct = res.limit > 0 ? (res.used / res.limit) * 100 : 0;
          const statusClass = getStatusClass(pct);
          return (
            <div key={res.categoryId}>
              <div className="stat-label">{res.name}用量</div>
              <div className="stat-value" style={{ fontSize: 20 }}>
                {formatNumber(res.used)}{" "}
                <span style={{ fontSize: 13, color: "var(--muted)" }}>
                  / {formatNumber(res.limit)} {res.unit}
                </span>
              </div>
              <div className="progress-track">
                <div
                  className={`progress-fill ${statusClass}`}
                  style={{ width: `${Math.min(pct, 100)}%` }}
                />
              </div>
            </div>
          );
        })}
      </div>
      {projects.length > 0 && (
        <div style={{ marginTop: 12, paddingTop: 10, borderTop: "1px solid var(--border)", fontSize: 12, color: "var(--muted)" }}>
          <svg style={{ verticalAlign: -2, marginRight: 4 }} width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <rect x="2" y="3" width="12" height="10" rx="2" />
            <path d="M2 6h12" />
          </svg>
          {projects.length} 个项目 · {projects.map((p) => p.name).join(", ")}
        </div>
      )}
    </div>
  );
};
