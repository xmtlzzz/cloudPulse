import React from "react";

interface SidebarProps {
  activeScreen: string;
  onNavigate: (screen: string) => void;
}

const services = [
  { id: "detail-vercel", name: "Vercel", icon: "vercel" },
  { id: "detail-neon", name: "Neon", icon: "neon" },
  { id: "detail-supabase", name: "Supabase", icon: "supabase" },
  { id: "detail-cloudflare", name: "Cloudflare", icon: "cloudflare" },
];

export const Sidebar: React.FC<SidebarProps> = ({ activeScreen, onNavigate }) => {
  return (
    <nav className="sidebar">
      <div className="sidebar-brand">
        <div className="logo">CP</div>
        <span className="name">CloudPulse</span>
      </div>

      <button
        className={`sidebar-item ${activeScreen === "dashboard" ? "active" : ""}`}
        onClick={() => onNavigate("dashboard")}
      >
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <rect x="2" y="2" width="5" height="5" rx="1" />
          <rect x="9" y="2" width="5" height="5" rx="1" />
          <rect x="2" y="9" width="5" height="5" rx="1" />
          <rect x="9" y="9" width="5" height="5" rx="1" />
        </svg>
        仪表盘
      </button>

      <div className="sidebar-section-label">服务</div>
      {services.map((svc) => (
        <button
          key={svc.id}
          className={`sidebar-item ${activeScreen === svc.id ? "active" : ""}`}
          onClick={() => onNavigate(svc.id)}
        >
          <ServiceIconSVG icon={svc.icon} />
          {svc.name}
        </button>
      ))}

      <div className="sidebar-spacer" />
      <div className="sidebar-footer">
        <button
          className={`sidebar-item ${activeScreen === "settings" ? "active" : ""}`}
          onClick={() => onNavigate("settings")}
        >
          <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="8" cy="8" r="2.5" />
            <path d="M8 1.5v2M8 12.5v2M1.5 8h2M12.5 8h2M3.4 3.4l1.4 1.4M11.2 11.2l1.4 1.4M3.4 12.6l1.4-1.4M11.2 4.8l1.4-1.4" />
          </svg>
          设置
        </button>
      </div>
    </nav>
  );
};

const ServiceIconSVG: React.FC<{ icon: string }> = ({ icon }) => {
  switch (icon) {
    case "vercel":
      return (
        <svg viewBox="0 0 16 16" fill="currentColor">
          <polygon points="8,2 14,14 2,14" />
        </svg>
      );
    case "neon":
      return (
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <circle cx="8" cy="8" r="5" />
          <path d="M8 3v5l3 3" />
        </svg>
      );
    case "supabase":
      return (
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M4 12V4l8 8V4" />
        </svg>
      );
    case "cloudflare":
      return (
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
          <path d="M3 10a4 4 0 018 0" />
          <circle cx="8" cy="6" r="2" />
        </svg>
      );
    default:
      return null;
  }
};
