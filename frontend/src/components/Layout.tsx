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
import { api } from '../services/api'

const { Content, Sider } = AntLayout

interface LayoutProps {
  children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const location = useLocation()
  const [checkingUpdate, setCheckingUpdate] = useState(false)

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
        message.success(`发现新版本 ${versionInfo.latestVersion}，请前往首页查看详情`)
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
    </AntLayout>
  )
}

export default Layout
