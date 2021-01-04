import { combineReducers } from "redux";
import anywhere from "./anywhere";
import anywhereAdd from "./anywhereAdd";
import note from "./note";
import user from "./user";

const rootReducer = combineReducers({
  anywhere,
  note,
  anywhereAdd,
  user
});

export default rootReducer;
