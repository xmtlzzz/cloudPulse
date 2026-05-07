import React, { useState } from "react";

interface TopbarProps {
  title: string;
  onSettingsClick: () => void;
  onThemeToggle: () => void;
  onSync: () => Promise<void>;
  isDark: boolean;
}

export const Topbar: React.FC<TopbarProps> = ({ title, onSettingsClick, onThemeToggle, onSync, isDark }) => {
  const [syncing, setSyncing] = useState(false);
  const [synced, setSynced] = useState(false);

  const handleSync = async () => {
    if (syncing) return;
    setSyncing(true);
    try {
      await onSync();
      setSyncing(false);
      setSynced(true);
      setTimeout(() => setSynced(false), 2000);
    } catch (e) {
      setSyncing(false);
    }
  };

  return (
    <div className="topbar">
      <div className="topbar-left">
        <span className="topbar-title">{title}</span>
      </div>
      <div className="topbar-right">
        <button
          className={`btn btn-sm btn-sync ${syncing ? "syncing" : ""}`}
          onClick={handleSync}
          disabled={syncing}
          style={synced ? { color: "var(--success)", borderColor: "var(--success)" } : {}}
        >
          {synced ? (
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M3 8l4 4 6-8" />
            </svg>
          ) : (
            <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M2 8a6 6 0 0111.5-2.3M14 8a6 6 0 01-11.5 2.3" />
              <path d="M14 2v4h-4M2 14v-4h4" />
            </svg>
          )}
          {synced ? "已同步" : "同步全部"}
        </button>
        <button className="btn btn-sm" onClick={onSettingsClick}>
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M2 4h12M2 8h12M2 12h12" />
          </svg>
          管理 Token
        </button>
        <button className="theme-toggle" onClick={onThemeToggle} title="切换明暗主题">
          {isDark ? (
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <path d="M13.5 8.5a5.5 5.5 0 01-7-7 5.5 5.5 0 107 7z" />
            </svg>
          ) : (
            <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="8" cy="8" r="3" />
              <path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.1 3.1l1.4 1.4M11.5 11.5l1.4 1.4M3.1 12.9l1.4-1.4M11.5 4.5l1.4-1.4" />
            </svg>
          )}
        </button>
      </div>
    </div>
  );
};
