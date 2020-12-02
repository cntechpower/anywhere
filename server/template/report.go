package template

const HTMLReportCss = `
  table.minimalistBlack {
    border: 3px solid #000000;
    width: 100%;
    text-align: left;
    border-collapse: collapse;
  }
  table.minimalistBlack td,
  table.minimalistBlack th {
    border: 1px solid #000000;
    padding: 5px 4px;
  }
  table.minimalistBlack tbody td {
    font-size: 13px;
  }
  table.minimalistBlack thead {
    background: #cfcfcf;
    background: -moz-linear-gradient(
      top,
      #dbdbdb 0%,
      #d3d3d3 66%,
      #cfcfcf 100%
    );
    background: -webkit-linear-gradient(
      top,
      #dbdbdb 0%,
      #d3d3d3 66%,
      #cfcfcf 100%
    );
    background: linear-gradient(
      to bottom,
      #dbdbdb 0%,
      #d3d3d3 66%,
      #cfcfcf 100%
    );
    border-bottom: 3px solid #000000;
  }
  table.minimalistBlack thead th {
    font-size: 15px;
    font-weight: bold;
    color: #000000;
    text-align: left;
  }
  table.minimalistBlack tfoot {
    font-size: 14px;
    font-weight: bold;
    color: #000000;
    border-top: 3px solid #000000;
  }
  table.minimalistBlack tfoot td {
    font-size: 14px;
  }`
const HTMLReport = `
<style>
%v
</style>
<h2>防火墙信息</h2>
%v
<h2>路由配置</h2>
%v
`
