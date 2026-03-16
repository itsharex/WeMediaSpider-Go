import React, { useState, useEffect, useRef, useCallback } from 'react'
import { Layout as AntLayout, Menu, Button, App } from 'antd'
import { useNavigate, useLocation } from 'react-router-dom'
import {
  HomeOutlined,
  CloudDownloadOutlined,
  FileTextOutlined,
  BarChartOutlined,
  ClockCircleOutlined,
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
  const [contentZoom, setContentZoom] = useState(1)
  const contentRef = useRef<HTMLDivElement>(null)

  // 基准设计尺寸（1100x700 窗口去掉 50px 侧栏、40px 标题栏）
  const BASE_WIDTH = 1050
  const BASE_HEIGHT = 660

  const updateZoom = useCallback(() => {
    if (!contentRef.current) return
    const parent = contentRef.current.parentElement
    if (!parent) return
    const availW = parent.clientWidth
    const availH = parent.clientHeight
    if (availW <= 0 || availH <= 0) return
    const zoom = Math.min(availW / BASE_WIDTH, availH / BASE_HEIGHT)
    setContentZoom(zoom)
  }, [])

  useEffect(() => {
    updateZoom()
    window.addEventListener('resize', updateZoom)
    return () => window.removeEventListener('resize', updateZoom)
  }, [updateZoom])

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
    {
      key: '/analytics',
      icon: <BarChartOutlined />,
      label: '分析',
    },
    {
      key: '/schedule',
      icon: <ClockCircleOutlined />,
      label: '定时',
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
    <AntLayout style={{ height: '100vh', display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
      {/* 自定义标题栏 */}
      <TitleBar />

      <AntLayout style={{ flex: 1, minHeight: 0 }}>
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
        <AntLayout style={{ overflow: 'hidden', minHeight: 0 }}>
          <Content
            style={{
              padding: 0,
              background: 'transparent',
              overflow: 'hidden',
              height: '100%',
            }}
          >
            <div
              ref={contentRef}
              style={{
                width: `${100 / contentZoom}%`,
                height: `${100 / contentZoom}%`,
                transform: `scale(${contentZoom})`,
                transformOrigin: 'top left',
                overflow: 'hidden',
                padding: '12px',
              }}
            >
              <PageTransition>
                {children}
              </PageTransition>
            </div>
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
