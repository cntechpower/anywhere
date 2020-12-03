package anywhereServer

import (
	"fmt"
	"strconv"
	"strings"

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
	_, err := cronTab.AddFunc(conf.Conf.ReportCron, s.SendDailyReport)
	if err != nil {
		panic(err)
	}
	cronTab.Start()
}

func (s *Server) SendDailyReport() {
	h := log.NewHeader("sendDailyReport")
	proxy, err := s.GetProxyConfigHtmlReport(10)
	if err != nil {
		log.Errorf(h, "get proxy html error: %v", err)
		return
	}
	whiteList, err := s.GetWhiteListHtmlReport()
	if err != nil {
		//do not return because we still need send proxy report.
		log.Errorf(h, "get whiteList html error: %v", err)
	}
	if err := tool.Send([]string{"root@cntechpower.com"}, "Anywhere Daily Report",
		fmt.Sprintf(template.HTMLReport, template.HTMLReportCss, whiteList, proxy)); err != nil {
		log.Errorf(h, "send mail error: %v", err)
	}
}

func (s *Server) GetProxyConfigHtmlReport(maxLines int) (html string, err error) {
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

func (s *Server) GetWhiteListHtmlReport() (html string, err error) {
	proxyDenyHtmlTable := strings.Builder{}
	totalCount := int64(0)
	proxyDenys, err := persist.GetTotalDenyRank()
	if err != nil {
		proxyDenyHtmlTable.WriteString(fmt.Sprintf("<b>%v!</b>", err))
		return proxyDenyHtmlTable.String(), err
	}
	proxyDenyHtmlTable.WriteString(`
<table class="minimalistBlack">
  <thead>
    <tr>
      <th>IP</th>
      <th>拒绝次数</th>
    </tr>
  </thead>
  <tbody>`)

	for _, p := range proxyDenys {
		totalCount += p.Count
		proxyDenyHtmlTable.WriteString(fmt.Sprintf(`
    <tr>
      <td>%v</td>
      <td>%v</td>
    </tr>`,
			p.Ip, p.Count))
	}
	proxyDenyHtmlTable.WriteString(fmt.Sprintf(`
    <tfoot>
        <tr>
          <td>总数:%v </td>
          <td>总数:%v</td>
        </tr>
      </tfoot>
   </tbody>
</table>`,
		len(proxyDenys), totalCount))
	return proxyDenyHtmlTable.String(), nil
}
