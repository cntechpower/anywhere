import React from "react";
import PropTypes from "prop-types";
import { Modal, Button } from "antd";

const { confirm } = Modal;

class ButtonWithConfirm extends React.Component {
  showConfirm(props) {
    confirm({
      title: props.confirmTitle,
      content: props.confirmContent,
      onOk() {
        props.fnOnOk();
      },
      onCancel() {}
    });
  }

  render() {
    return (
      <Button
        disabled={this.props.btnDisabled}
        onClick={() => this.showConfirm(this.props)}
      >
        {this.props.btnName}
      </Button>
    );
  }
}

ButtonWithConfirm.propTypes = {
  btnName: PropTypes.string.isRequired,
  btnDisabled: PropTypes.bool.isRequired,
  confirmTitle: PropTypes.string.isRequired,
  confirmContent: PropTypes.string.isRequired,
  fnOnOk: PropTypes.func.isRequired
};

export default ButtonWithConfirm;
