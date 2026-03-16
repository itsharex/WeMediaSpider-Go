import React, { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider, theme, App as AntApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import Layout from './components/Layout'
import LogFloatButton from './components/LogFloatButton'
import HomePage from './pages/HomePage'
import ScrapePage from './pages/ScrapePage'
import ResultsPage from './pages/ResultsPage'
import AnalyticsPage from './pages/AnalyticsPage'
import SchedulePage from './pages/SchedulePage'
import SettingsPage from './pages/SettingsPage'
import './App.css'

function App() {
  // 所有数值输入框 focus 时自动全选
  useEffect(() => {
    const handleFocus = (e: FocusEvent) => {
      const target = e.target as HTMLInputElement
      if (target.tagName === 'INPUT' && (target.type === 'number' || target.classList.contains('ant-input-number-input'))) {
        requestAnimationFrame(() => target.select())
      }
    }
    document.addEventListener('focus', handleFocus, true)
    return () => document.removeEventListener('focus', handleFocus, true)
  }, [])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#07C160',
          colorBgBase: '#1a1a1a',
          borderRadius: 4,
          fontSize: 13,
          controlHeight: 30,
        },
        components: {
          Card: {
            paddingLG: 16,
            headerHeight: 40,
          },
          Form: {
            itemMarginBottom: 12,
            labelHeight: 20,
          },
          Table: {
            cellPaddingBlock: 6,
            headerBg: '#141414',
          },
          Button: {
            controlHeight: 30,
            paddingContentHorizontal: 12,
          },
          Input: {
            controlHeight: 30,
            paddingBlock: 4,
          },
          Select: {
            controlHeight: 30,
          },
        },
      }}
    >
      <AntApp>
        <BrowserRouter>
          <Layout>
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/home" element={<HomePage />} />
              <Route path="/scrape" element={<ScrapePage />} />
              <Route path="/results" element={<ResultsPage />} />
              <Route path="/analytics" element={<AnalyticsPage />} />
              <Route path="/schedule" element={<SchedulePage />} />
              <Route path="/settings" element={<SettingsPage />} />
            </Routes>
          </Layout>
          <LogFloatButton />
        </BrowserRouter>
      </AntApp>
    </ConfigProvider>
  )
}

export default App
