import {
  NOTE_REFRESH_NOTE_LIST,
  NOTE_SET_ERROR
} from "../constants/ActionTypes";

import api from "../api/apiManager";
export const refreshNoteList = data => ({
  type: NOTE_REFRESH_NOTE_LIST,
  data
});

export const setError = error => ({
  type: NOTE_SET_ERROR,
  error
});

export const fetchNoteList = () => {
  return function(dispatch) {
    return fetch(api.noteListApi)
      .then(response => response.json())
      .then(
        response => {
          dispatch(refreshNoteList(response));
        },
        error => {
          dispatch(setError(error));
        }
      );
  };
};
