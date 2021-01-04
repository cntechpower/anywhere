import { connect } from "react-redux";
import * as UserActions from "../actions/user";
import { bindActionCreators } from "redux";
import MainSection from "../components/User/Login";

function mapStateToProps(state) {
  return {
    init: state.user.init,
    loading: state.user.loading,
    isLogin: state.user.isLogin,
    isLoginError: state.user.isLoginError,
    error: state.user.error
  };
}

const mapDispatchToProps = dispatch => ({
  actions: bindActionCreators(UserActions, dispatch)
});

export default connect(mapStateToProps, mapDispatchToProps)(MainSection);
