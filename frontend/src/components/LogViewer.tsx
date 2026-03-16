import React, { useState, useEffect, useRef } from 'react';
import { Drawer, Button, Space } from 'antd';
import { GetAllLogs, ClearLogs } from '../../wailsjs/go/app/App';
import './LogViewer.css';

interface LogViewerProps {
  visible: boolean;
  onClose: () => void;
}

const LogViewer: React.FC<LogViewerProps> = ({ visible, onClose }) => {
  const [logs, setLogs] = useState<string[]>([]);
  const [autoScroll, setAutoScroll] = useState(true);
  const logContainerRef = useRef<HTMLDivElement>(null);

  const loadLogs = async () => {
    try {
      const result = await GetAllLogs();
      setLogs(result || []);
    } catch (error) {
      console.error('加载日志失败:', error);
    }
  };

  const handleClear = async () => {
    try {
      await ClearLogs();
      setLogs([]);
    } catch (error) {
      console.error('清空日志失败:', error);
    }
  };

  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  useEffect(() => {
    if (visible) {
      loadLogs();
      const interval = setInterval(loadLogs, 1000);
      return () => clearInterval(interval);
    }
  }, [visible]);

  return (
    <Drawer
      title="运行日志"
      placement="right"
      open={visible}
      onClose={onClose}
      width={400}
      mask={false}
      closable={false}
      extra={
        <Space>
          <Button size="small" onClick={() => setAutoScroll(!autoScroll)}>
            {autoScroll ? '停止滚动' : '自动滚动'}
          </Button>
          <Button size="small" onClick={handleClear}>清空</Button>
          <Button size="small" onClick={loadLogs}>刷新</Button>
        </Space>
      }
    >
      <div ref={logContainerRef} className="log-container">
        {logs.length === 0 ? (
          <div className="log-empty">暂无日志</div>
        ) : (
          logs.map((log, index) => (
            <div key={index} className="log-line">
              {log}
            </div>
          ))
        )}
      </div>
    </Drawer>
  );
};

export default LogViewer;
