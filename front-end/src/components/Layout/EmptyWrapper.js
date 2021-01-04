import React from "react";

class EmptyWrapper extends React.Component {
  render() {
    const divStyle = {
      color: "blue",
      minHeight: "100vh"
      // backgroundImage:
      //   "url(http://10.0.0.2/img/anywhere_login_backgroud.jpm.jpg)",
      // backgroundRepeat: "no-repeat",
      // backgroundSize: "cover"
    };
    return <div style={divStyle}>{this.props.children}</div>;
  }
}

export default EmptyWrapper;
