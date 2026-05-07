import React from "react";

interface SyncResult {
  platform: string;
  success: boolean;
  error?: string;
}

interface SyncResultModalProps {
  results: SyncResult[];
  onClose: () => void;
}

export const SyncResultModal: React.FC<SyncResultModalProps> = ({ results, onClose }) => {
  const successCount = results.filter((r) => r.success).length;
  const failCount = results.filter((r) => !r.success).length;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 200,
      }}
    >
      <div className="card" style={{ width: 420 }}>
        <div className="card-header">
          <span className="card-title">同步结果</span>
          <button className="btn btn-sm" onClick={onClose}>
            关闭
          </button>
        </div>

        <div style={{ display: "flex", gap: 16, marginBottom: 20 }}>
          <div className="card" style={{ flex: 1, textAlign: "center", padding: 16 }}>
            <div className="stat-value" style={{ fontSize: 24, color: "var(--success)" }}>
              {successCount}
            </div>
            <div className="stat-label">成功</div>
          </div>
          <div className="card" style={{ flex: 1, textAlign: "center", padding: 16 }}>
            <div className="stat-value" style={{ fontSize: 24, color: failCount > 0 ? "var(--danger)" : "var(--muted)" }}>
              {failCount}
            </div>
            <div className="stat-label">失败</div>
          </div>
        </div>

        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {results.map((result, i) => (
            <div
              key={i}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                padding: "10px 12px",
                background: "var(--bg)",
                borderRadius: 6,
              }}
            >
              <span style={{ fontWeight: 500, fontSize: 13 }}>{result.platform}</span>
              {result.success ? (
                <span className="pill pill-success">
                  <span className="pill-dot" />
                  成功
                </span>
              ) : (
                <span className="pill pill-danger" title={result.error}>
                  <span className="pill-dot" />
                  失败
                </span>
              )}
            </div>
          ))}
        </div>

        <div style={{ marginTop: 20, display: "flex", justifyContent: "flex-end" }}>
          <button className="btn btn-primary" onClick={onClose}>
            确定
          </button>
        </div>
      </div>
    </div>
  );
};
