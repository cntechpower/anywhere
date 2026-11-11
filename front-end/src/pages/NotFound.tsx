import { Button } from "antd";
import { CompassOutlined } from "@ant-design/icons";
import { Link } from "react-router-dom";
import EmptyState from "../components/EmptyState";

export default function NotFound() {
  return (
    <div className="panel">
      <EmptyState
        icon={<CompassOutlined />}
        title="页面不存在"
        description="你访问的地址可能已经被移动或删除。"
        action={
          <Link to="/home">
            <Button type="primary">返回概览</Button>
          </Link>
        }
      />
    </div>
  );
}
