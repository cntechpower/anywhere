import {
  NOTE_REFRESH_NOTE_LIST,
  NOTE_SET_ERROR
} from "../constants/ActionTypes";

const initialState = {
  data: [""],
  loading: true,
  error: null
};

export default function note(state = initialState, action) {
  switch (action.type) {
    case NOTE_REFRESH_NOTE_LIST:
      return Object.assign({}, state, {
        loading: false,
        error: null,
        data: action.data
      });
    case NOTE_SET_ERROR:
      return Object.assign({}, state, {
        loading: false,
        error: state.error.message
      });
    default:
      return state;
  }
}
