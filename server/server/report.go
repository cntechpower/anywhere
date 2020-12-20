package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cntechpower/anywhere/constants"

	"github.com/cntechpower/anywhere/server/conf"

	"github.com/robfig/cron/v3"

	"github.com/cntechpower/anywhere/log"
	"github.com/cntechpower/anywhere/server/persist"
	"github.com/cntechpower/anywhere/server/template"
	"github.com/cntechpower/anywhere/server/tool"
	"github.com/cntechpower/anywhere/util"
)

var cronTab *cron.Cron

func (s *Server) StartReportCron() {
	cronTab = cron.New()
	cron.WithLogger(log.NewHeader("cron"))
	_, err := cronTab.AddFunc(conf.Conf.ReportCron, s.SendDailyReport)
	if err != nil {
		panic(err)
	}
	cronTab.Start()
}

func (s *Server) SendDailyReport() {
	h := log.NewHeader("sendDailyReport")
	report, err := s.GetHtmlReport(h)
	if err != nil {
		return
	}
	if err := tool.Send([]string{"root@cntechpower.com"}, "Anywhere Daily Report", report); err != nil {
		log.Errorf(h, "send mail error: %v", err)
	}
}

func (s *Server) GetHtmlReport(h *log.Header) (string, error) {

	totalWhiteList, err := s.getWhiteListHtmlReport(persist.RankTypeTotal, 10)
	if err != nil {
		//do not return because we still need send proxy report.
		h.Errorf("get totalWhiteList html error: %v", err)
	}
	dailyWhiteList, err := s.getWhiteListHtmlReport(persist.RankTypeDaily, 10)
	if err != nil {
		h.Errorf("get dailyWhiteList html error: %v", err)
	}
	agent, err := s.getAgentsHtmlReport(10)
	if err != nil {
		h.Errorf("get proxy html error: %v", err)
	}
	proxy, err := s.getProxyConfigHtmlReport(10)
	if err != nil {
		h.Errorf("get proxy html error: %v", err)
	}
	return fmt.Sprintf(template.HTMLReport, template.HTMLReportCss, dailyWhiteList, totalWhiteList, agent, proxy), nil
}

func (s *Server) getProxyConfigHtmlReport(maxLines int) (html string, err error) {
	configs := s.ListProxyConfigs()

	configsHtmlTable := strings.Builder{}
	configsHtmlTable.WriteString(`
<table class="minimalistBlack">
  <thead>
    <tr>
      <th>节点名</th>
      <th>外网端口</th>
      <th>内网地址</th>
      <th>白名单开关</th>
      <th>流量</th>
    </tr>
  </thead>
  <tbody>`)
	totalFlows := float64(0)
	whiteListDisable := 0
	nodeMap := make(map[string]struct{}, 0)
	for idx, config := range configs {
		flows := float64(config.NetworkFlowRemoteToLocalInBytes+config.NetworkFlowLocalToRemoteInBytes) / 1024 / 1024
		totalFlows += flows
		nodeMap[config.AgentId] = struct{}{}
		if !config.IsWhiteListOn {
			whiteListDisable++
		}
		if idx < maxLines {
			configsHtmlTable.WriteString(fmt.Sprintf(`
    <tr>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%vMB</td>
    </tr>`,
				config.AgentId, config.RemotePort, config.LocalAddr, util.BoolToString(config.IsWhiteListOn),
				strconv.FormatFloat(flows, 'f', 5, 64)))
		}
	}
	configsHtmlTable.WriteString(fmt.Sprintf(`
    <tfoot>
        <tr>
          <td>节点数:%v </td>
          <td>--</td>
          <td>--</td>
          <td>未开启: %v</td>
          <td>总流量: %vMB</td>
        </tr>
      </tfoot>
   </tbody>
</table>`,
		len(nodeMap), whiteListDisable, strconv.FormatFloat(totalFlows, 'f', 5, 64)))
	return configsHtmlTable.String(), nil
}

func (s *Server) getWhiteListHtmlReport(typ string, maxLines int64) (html string, err error) {
	proxyDenyHtmlTable := strings.Builder{}
	denyDetails, detailCount, ipCount, err := persist.GetWhiteListDenyRank(typ, maxLines)
	if err != nil {
		proxyDenyHtmlTable.WriteString(fmt.Sprintf("<b>%v!</b>", err))
		return proxyDenyHtmlTable.String(), err
	}
	proxyDenyHtmlTable.WriteString(`
<table class="minimalistBlack">
  <thead>
    <tr>
      <th>IP</th>
      <th>所属地区</th>
      <th>拒绝次数</th>
    </tr>
  </thead>
  <tbody>`)

	for _, p := range denyDetails {
		proxyDenyHtmlTable.WriteString(fmt.Sprintf(`
    <tr>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
    </tr>`,
			p.Ip, p.Address, p.Count))

	}
	proxyDenyHtmlTable.WriteString(fmt.Sprintf(`
    <tfoot>
        <tr>
          <td>总数:%v </td>
          <td>--</td>
          <td>总数:%v</td>
        </tr>
      </tfoot>
   </tbody>
</table>`,
		ipCount, detailCount))
	return proxyDenyHtmlTable.String(), nil
}

func (s *Server) getAgentsHtmlReport(maxLines int) (html string, err error) {
	agents := s.ListAgentInfo()

	configsHtmlTable := strings.Builder{}
	configsHtmlTable.WriteString(`
<table class="minimalistBlack">
  <thead>
    <tr>
      <th>用户名</th>
      <th>节点ID</th>
      <th>内网地址</th>
      <th>配置总数</th>
      <th>心跳发送时间</th>
      <th>心跳接收时间</th>
      <th>延迟</th>
    </tr>
  </thead>
  <tbody>`)

	for idx, agent := range agents {
		if idx < maxLines {
			configsHtmlTable.WriteString(fmt.Sprintf(`
    <tr>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%v</td>
      <td>%vms</td>
    </tr>`,
				agent.UserName, agent.Id, agent.RemoteAddr,
				agent.ProxyConfigCount,
				agent.LastAckSend.Format(constants.DefaultTimeFormat),
				agent.LastAckRcv.Format(constants.DefaultTimeFormat),
				agent.LastAckRcv.Sub(agent.LastAckSend).Milliseconds()))
		}
	}
	configsHtmlTable.WriteString(fmt.Sprintf(`
    <tfoot>
        <tr>
          <td>节点数:%v </td>
          <td>--</td>
          <td>--</td>
          <td>--</td>
          <td>--</td>
          <td>--</td>
          <td>--</td>
        </tr>
      </tfoot>
   </tbody>
</table>`,
		len(agents)))
	return configsHtmlTable.String(), nil
}
