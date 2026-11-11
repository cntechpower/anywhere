import { Button } from "antd";
import { ReloadOutlined, WarningOutlined } from "@ant-design/icons";
import { apiErrorMessage } from "../utils/format";

interface ErrorStateProps {
  error: unknown;
  onRetry?: () => void;
  title?: string;
}

export default function ErrorState({
  error,
  onRetry,
  title = "加载失败",
}: ErrorStateProps) {
  return (
    <div className="state-block" role="alert">
      <span className="state-icon" aria-hidden="true">
        <WarningOutlined />
      </span>
      <h3>{title}</h3>
      <p>{apiErrorMessage(error)}</p>
      {onRetry && (
        <Button icon={<ReloadOutlined />} onClick={onRetry}>
          重新加载
        </Button>
      )}
    </div>
  );
}
