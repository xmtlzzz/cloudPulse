import React, { useState } from "react";
import { AuthConfig } from "../types";

interface AddServiceModalProps {
  platform: string;
  authConfig: AuthConfig | null;
  onSubmit: (accountName: string, credentials: Record<string, string>) => Promise<void>;
  onCancel: () => void;
}

export const AddServiceModal: React.FC<AddServiceModalProps> = ({
  platform,
  authConfig,
  onSubmit,
  onCancel,
}) => {
  const [accountName, setAccountName] = useState("");
  const [credentials, setCredentials] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCredentialChange = (fieldId: string, value: string) => {
    setCredentials((prev) => ({ ...prev, [fieldId]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!accountName.trim() || loading) return;
    setLoading(true);
    setError(null);
    try {
      await onSubmit(accountName, credentials);
    } catch (e: any) {
      setError(e.message || "添加失败，请检查凭证是否正确");
    } finally {
      setLoading(false);
    }
  };

  if (!authConfig) return null;

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
      <div
        className="card"
        style={{ width: 420, maxHeight: "80vh", overflow: "auto" }}
      >
        <div className="card-header">
          <span className="card-title">添加 {platform} 服务</span>
        </div>
        <form onSubmit={handleSubmit}>
          <div className="input-group">
            <label className="input-label">账户名称</label>
            <input
              className="input"
              type="text"
              placeholder="例如: Personal, Production"
              value={accountName}
              onChange={(e) => setAccountName(e.target.value)}
              required
            />
          </div>

          {authConfig.fields.map((field) => (
            <div className="input-group" key={field.id}>
              <label className="input-label">{field.name}</label>
              <input
                className="input"
                type={field.type}
                placeholder={field.description}
                value={credentials[field.id] || ""}
                onChange={(e) => handleCredentialChange(field.id, e.target.value)}
                required={field.required}
              />
            </div>
          ))}

          {error && (
            <div
              style={{
                padding: "8px 12px",
                fontSize: 12,
                color: "var(--danger)",
                background: "var(--danger-subtle)",
                borderRadius: 4,
                marginTop: 12,
              }}
            >
              {error}
            </div>
          )}

          <div style={{ display: "flex", gap: 8, justifyContent: "flex-end", marginTop: 20 }}>
            <button type="button" className="btn" onClick={onCancel} disabled={loading}>
              取消
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? "验证中..." : "添加"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
