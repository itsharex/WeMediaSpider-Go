import React, { useState } from 'react';
import { FloatButton } from 'antd';
import { FileTextOutlined } from '@ant-design/icons';
import LogViewer from './LogViewer';

const LogFloatButton: React.FC = () => {
  const [visible, setVisible] = useState(false);

  return (
    <>
      <FloatButton
        icon={<FileTextOutlined />}
        tooltip={visible ? '关闭日志' : '查看日志'}
        type={visible ? 'primary' : 'default'}
        onClick={() => setVisible(!visible)}
        style={{ right: 24, bottom: 24, zIndex: 1001 }}
      />
      <LogViewer visible={visible} onClose={() => setVisible(false)} />
    </>
  );
};

export default LogFloatButton;
