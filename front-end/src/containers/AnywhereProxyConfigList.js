import { connect } from "react-redux";
import * as AnywhereActions from "../actions/anywhere";
import { bindActionCreators } from "redux";
import MainSection from "../components/Anywhere/ProxyConfigList";

function mapStateToProps(state) {
  return {
    error: state.anywhere.error,
    loading: state.anywhere.loading,
    data: state.anywhere.data
  };
}

const mapDispatchToProps = dispatch => ({
  actions: bindActionCreators(AnywhereActions, dispatch)
});

export default connect(mapStateToProps, mapDispatchToProps)(MainSection);
