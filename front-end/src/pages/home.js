import React from "react";
import "antd/dist/antd.css";
class Home extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      error: null,
      isLoaded: false,
      status: null,
    };
  }
  render() {
    return "home";
  }
}

export default Home;
