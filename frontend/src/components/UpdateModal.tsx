import React, { useState } from 'react'
import { Modal, Button, Checkbox } from 'antd'

interface VersionInfo {
  currentVersion: string
  latestVersion: string
  hasUpdate: boolean
  updateUrl: string
  releaseNotes: string
}

interface UpdateModalProps {
  open: boolean
  versionInfo: VersionInfo | null
  onDownload: () => void
  onLater: () => void
  onIgnoreToday: () => void
}

const UpdateModal: React.FC<UpdateModalProps> = ({
  open,
  versionInfo,
  onDownload,
  onLater,
  onIgnoreToday,
}) => {
  const [ignoreToday, setIgnoreToday] = useState(false)

  // 当对话框关闭时重置复选框状态
  React.useEffect(() => {
    if (!open) {
      setIgnoreToday(false)
    }
  }, [open])

  if (!versionInfo) return null

  const handleLater = () => {
    console.log('[UpdateModal] handleLater called, ignoreToday=', ignoreToday)
    if (ignoreToday) {
      console.log('[UpdateModal] Calling onIgnoreToday')
      onIgnoreToday()
    } else {
      console.log('[UpdateModal] Calling onLater')
      onLater()
    }
  }

  return (
    <Modal
      open={open}
      onCancel={handleLater}
      footer={null}
      width={320}
      centered
      closable={true}
      styles={{
        body: { padding: '20px' }
      }}
    >
      {/* 图标和标题 */}
      <div style={{
        textAlign: 'center',
        marginBottom: 16,
      }}>
        <div style={{
          fontSize: 32,
          marginBottom: 8,
        }}>
          ✨
        </div>
        <div style={{
          fontSize: 16,
          fontWeight: 500,
          color: '#000',
        }}>
          新版本 {versionInfo.latestVersion}
        </div>
      </div>

      {/* 更新说明 */}
      {versionInfo.releaseNotes && (
        <div style={{
          fontSize: 13,
          color: '#666',
          lineHeight: 1.6,
          marginBottom: 16,
          textAlign: 'center',
        }}>
          {versionInfo.releaseNotes}
        </div>
      )}

      {/* 今日不再提示 */}
      <div style={{
        marginBottom: 16,
        textAlign: 'center',
      }}>
        <Checkbox
          checked={ignoreToday}
          onChange={(e) => setIgnoreToday(e.target.checked)}
        >
          <span style={{ fontSize: 12, color: '#999' }}>
            今日不再提示
          </span>
        </Checkbox>
      </div>

      {/* 按钮 */}
      <div style={{
        display: 'flex',
        gap: 8,
      }}>
        <Button
          block
          onClick={handleLater}
          style={{
            height: 36,
          }}
        >
          稍后
        </Button>
        <Button
          block
          type="primary"
          onClick={onDownload}
          style={{
            height: 36,
            background: '#07C160',
            borderColor: '#07C160',
          }}
        >
          立即更新
        </Button>
      </div>
    </Modal>
  )
}

export default UpdateModal
