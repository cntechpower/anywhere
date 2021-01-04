import { connect } from "react-redux";
import * as AnywhereAddActions from "../actions/anywhereAdd";
import { bindActionCreators } from "redux";
import MainSection from "../components/Anywhere/ProxyConfigAdd";

function mapStateToProps(state) {
  return {
    config: state.anywhereAdd.config,
    init: state.anywhereAdd.init,
    error: state.anywhereAdd.error,
    creating: state.anywhereAdd.creating,
    createdOk: state.anywhereAdd.createdOk,
    agentsLoading: state.anywhereAdd.agentsLoading,
    agents: state.anywhereAdd.agents,
    localIp: state.anywhereAdd.localIp
  };
}

const mapDispatchToProps = dispatch => ({
  actions: bindActionCreators(AnywhereAddActions, dispatch)
});

export default connect(mapStateToProps, mapDispatchToProps)(MainSection);
