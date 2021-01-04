import { connect } from "react-redux";
import * as NoteActions from "../actions/note";
import { bindActionCreators } from "redux";
import MainSection from "../components/Note/NoteList";

function mapStateToProps(state) {
  return {
    error: state.note.error,
    loading: state.note.loading,
    data: state.note.data
  };
}

const mapDispatchToProps = dispatch => ({
  actions: bindActionCreators(NoteActions, dispatch)
});

export default connect(mapStateToProps, mapDispatchToProps)(MainSection);
