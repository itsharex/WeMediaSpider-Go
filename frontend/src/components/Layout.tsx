import React, { useState } from 'react'
import { Layout as AntLayout, Menu, Button, App } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import {
  HomeOutlined,
  CloudDownloadOutlined,
  FileTextOutlined,
  SettingOutlined,
  WechatOutlined,
  SyncOutlined,
} from '@ant-design/icons'
import TitleBar from './TitleBar'
import PageTransition from './PageTransition'
import UpdateModal from './UpdateModal'
import { api } from '../services/api'

const { Content, Sider } = AntLayout

interface LayoutProps {
  children: React.ReactNode
}

interface VersionInfo {
  currentVersion: string
  latestVersion: string
  hasUpdate: boolean
  updateUrl: string
  releaseNotes: string
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const location = useLocation()
  const [checkingUpdate, setCheckingUpdate] = useState(false)
  const [showUpdateModal, setShowUpdateModal] = useState(false)
  const [updateInfo, setUpdateInfo] = useState<VersionInfo | null>(null)

  const menuItems = [
    {
      key: '/',
      icon: <HomeOutlined />,
      label: '首页',
    },
    {
      key: '/scrape',
      icon: <CloudDownloadOutlined />,
      label: '爬取',
    },
    {
      key: '/results',
      icon: <FileTextOutlined />,
      label: '数据',
    },
  ]

  // 检查更新
  const handleCheckUpdate = async () => {
    try {
      setCheckingUpdate(true)
      const versionInfo = await api.checkForUpdates()
      if (versionInfo.hasUpdate) {
        setUpdateInfo(versionInfo)
        setShowUpdateModal(true)
      } else {
        message.success('当前已是最新版本')
      }
    } catch (error) {
      console.error('检查更新失败:', error)
      message.error('检查更新失败，请稍后重试')
    } finally {
      setCheckingUpdate(false)
    }
  }

  // 立即下载
  const handleDownloadUpdate = () => {
    if (updateInfo?.updateUrl) {
      window.open(updateInfo.updateUrl, '_blank')
      setShowUpdateModal(false)
    }
  }

  // 稍后更新
  const handleLaterUpdate = () => {
    setShowUpdateModal(false)
  }

  // 今日不再提示
  const handleIgnoreToday = async () => {
    const today = new Date().toISOString().split('T')[0]
    console.log('[Layout] handleIgnoreToday called, today=', today)
    try {
      const { SetUpdateIgnoredDate } = await import('../../wailsjs/go/app/App')
      console.log('[Layout] Calling SetUpdateIgnoredDate with date:', today)
      await SetUpdateIgnoredDate(today)
      console.log('[Layout] SetUpdateIgnoredDate completed successfully')
      setShowUpdateModal(false)
      message.info('今日将不再提示更新')
    } catch (error) {
      console.error('[Layout] 设置忽略日期失败:', error)
      message.error('设置失败')
    }
  }

  return (
    <AntLayout style={{ minHeight: '100vh', maxHeight: '100vh', overflow: 'hidden' }}>
      {/* 自定义标题栏 */}
      <TitleBar />

      <AntLayout style={{ height: 'calc(100vh - 40px)' }}>
        <Sider
          theme="dark"
          width={50}
          collapsedWidth={50}
          collapsed={true}
          collapsible={false}
          style={{
            background: '#0d0d0d',
            overflow: 'hidden',
            position: 'relative',
          }}
        >
          <Menu
            theme="dark"
            mode="inline"
            selectedKeys={[location.pathname]}
            items={menuItems}
            onClick={({ key }) => navigate(key)}
            style={{
              background: '#0d0d0d',
              fontSize: 14,
              paddingTop: 8,
            }}
            inlineCollapsed={true}
          />

          {/* 设置按钮 - 绝对定位在底部第二个位置 */}
          <div
            onClick={() => navigate('/settings')}
            style={{
              position: 'absolute',
              bottom: 40,
              left: 0,
              right: 0,
              height: 40,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              color: location.pathname === '/settings' ? '#fff' : 'rgba(255, 255, 255, 0.65)',
              transition: 'all 0.3s',
              background: location.pathname === '/settings' ? 'rgba(7, 193, 96, 0.2)' : '#0d0d0d',
            }}
            onMouseEnter={(e) => {
              if (location.pathname !== '/settings') {
                e.currentTarget.style.background = 'rgba(255, 255, 255, 0.08)'
                e.currentTarget.style.color = 'rgba(255, 255, 255, 0.85)'
              }
            }}
            onMouseLeave={(e) => {
              if (location.pathname !== '/settings') {
                e.currentTarget.style.background = '#0d0d0d'
                e.currentTarget.style.color = 'rgba(255, 255, 255, 0.65)'
              }
            }}
          >
            <SettingOutlined style={{ fontSize: 16 }} />
          </div>

          {/* 检查更新按钮 - 绝对定位在最底部 */}
          <div
            onClick={handleCheckUpdate}
            style={{
              position: 'absolute',
              bottom: 0,
              left: 0,
              right: 0,
              height: 40,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              color: 'rgba(255, 255, 255, 0.65)',
              transition: 'all 0.3s',
              borderTop: '1px solid #1f1f1f',
              background: '#0d0d0d',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(255, 255, 255, 0.08)'
              e.currentTarget.style.color = 'rgba(255, 255, 255, 0.85)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = '#0d0d0d'
              e.currentTarget.style.color = 'rgba(255, 255, 255, 0.65)'
            }}
          >
            <SyncOutlined spin={checkingUpdate} style={{ fontSize: 16 }} />
          </div>
        </Sider>
        <AntLayout style={{ overflow: 'hidden' }}>
          <Content
            style={{
              padding: '12px',
              background: 'transparent',
              overflow: 'hidden',
              height: '100%',
            }}
          >
            <PageTransition>
              {children}
            </PageTransition>
          </Content>
        </AntLayout>
      </AntLayout>

      {/* 更新提示对话框 */}
      <UpdateModal
        open={showUpdateModal}
        versionInfo={updateInfo}
        onDownload={handleDownloadUpdate}
        onLater={handleLaterUpdate}
        onIgnoreToday={handleIgnoreToday}
      />
    </AntLayout>
  )
}

export default Layout
