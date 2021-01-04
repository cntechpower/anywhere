import React from "react";
import { Statistic, Row, Col } from "antd";

const Summary = () => (
  <Row gutter={26}>
    <Col span={12}>
      <Statistic title="Active Users" value={112893} />
    </Col>
    <Col span={12}>
      <Statistic title="Active Users" value={112893} />
    </Col>
    <Col span={12}>
      <Statistic title="Active Users" value={112893} />
    </Col>
    <Col span={12}>
      <Statistic title="Active Users" value={112893} />
    </Col>
  </Row>
);

export default Summary;
