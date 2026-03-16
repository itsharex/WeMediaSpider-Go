import React, { useState, useEffect, useRef } from 'react'
import {
  Card,
  Button,
  Space,
  Input,
  Form,
  DatePicker,
  Switch,
  Progress,
  Table,
  App,
  Alert,
  Select,
  Dropdown,
  Tag,
  Modal,
  Tooltip,
} from 'antd'
import { useNavigate } from 'react-router-dom'
import {
  PlayCircleOutlined,
  StopOutlined,
  CheckCircleOutlined,
  LoadingOutlined,
  CloseCircleOutlined,
  StarOutlined,
  StarFilled,
  PlusOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useScrapeStore } from '../stores/scrapeStore'
import { useLoginStore } from '../stores/loginStore'
import { useFormStore } from '../stores/formStore'
import { useConfigStore } from '../stores/configStore'
import { api, events } from '../services/api'
import type { ScrapeConfig, AccountStatus } from '../types'

const { TextArea } = Input
const { RangePicker } = DatePicker

const ScrapePage: React.FC = () => {
  const { message } = App.useApp()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const { loginStatus } = useLoginStore()
  const { config, setConfig } = useConfigStore()
  const formStore = useFormStore()
  const {
    articles,
    progress,
    accountStatuses,
    isScrapingInProgress,
    setArticles,
    setProgress,
    addAccountStatus,
    updateAccountStatus,
    clearAccountStatuses,
    setScrapingInProgress,
  } = useScrapeStore()

  const [loading, setLoading] = useState(false)
  const [totalArticleCount, setTotalArticleCount] = useState(0)
  const [accountArticleCounts, setAccountArticleCounts] = useState<Record<string, number>>({})
  const [favoriteAccounts, setFavoriteAccounts] = useState<string[]>([])
  const [selectedAccounts, setSelectedAccounts] = useState<string[]>([])
  const [addAccountModalVisible, setAddAccountModalVisible] = useState(false)
  const [newAccountName, setNewAccountName] = useState('')
  const [selectedCloudDrive, setSelectedCloudDrive] = useState<string | undefined>(undefined)
  const hasFreqLimitErrorRef = useRef(false) // 使用 ref 跟踪频率限制

  // 加载常用公众号
  useEffect(() => {
    const saved = localStorage.getItem('favoriteAccounts')
    if (saved) {
      setFavoriteAccounts(JSON.parse(saved))
    }
  }, [])

  // 初始化中国时间
  useEffect(() => {
    formStore.initChinaTime()
  }, [])

  // 保存常用公众号
  const saveFavoriteAccounts = (accounts: string[]) => {
    localStorage.setItem('favoriteAccounts', JSON.stringify(accounts))
    setFavoriteAccounts(accounts)
  }

  // 添加到常用公众号
  const addToFavorites = (accountName: string) => {
    if (!favoriteAccounts.includes(accountName)) {
      const newFavorites = [...favoriteAccounts, accountName]
      saveFavoriteAccounts(newFavorites)
      message.success(`已添加 "${accountName}" 到常用列表`)
    }
  }

  // 从常用公众号移除
  const removeFromFavorites = (accountName: string) => {
    const newFavorites = favoriteAccounts.filter(a => a !== accountName)
    saveFavoriteAccounts(newFavorites)
    message.success(`已从常用列表移除 "${accountName}"`)
  }

  // 切换公众号选中状态
  const toggleAccountSelection = (accountName: string) => {
    const currentAccounts = form.getFieldValue('accounts') || ''
    const accountList = currentAccounts.split('\n').filter((a: string) => a.trim())

    if (accountList.includes(accountName)) {
      // 移除
      const newList = accountList.filter((a: string) => a !== accountName)
      form.setFieldValue('accounts', newList.join('\n'))
      setSelectedAccounts(newList)
    } else {
      // 添加
      const newList = [...accountList, accountName]
      form.setFieldValue('accounts', newList.join('\n'))
      setSelectedAccounts(newList)
    }
    handleFormChange()
  }

  // 添加新的常用公众号（支持批量）
  const handleAddNewAccount = () => {
    if (!newAccountName.trim()) {
      message.warning('请输入公众号名称')
      return
    }

    // 按行分割，支持批量添加
    const accountsToAdd = newAccountName
      .split('\n')
      .map(name => name.trim())
      .filter(name => name.length > 0)

    if (accountsToAdd.length === 0) {
      message.warning('请输入公众号名称')
      return
    }

    // 过滤掉已存在的公众号
    const newAccounts = accountsToAdd.filter(account => !favoriteAccounts.includes(account))
    const duplicates = accountsToAdd.filter(account => favoriteAccounts.includes(account))

    if (newAccounts.length === 0) {
      message.warning('所有公众号都已在常用列表中')
      return
    }

    // 添加到常用列表
    const updatedFavorites = [...favoriteAccounts, ...newAccounts]
    saveFavoriteAccounts(updatedFavorites)

    // 显示结果
    if (duplicates.length > 0) {
      message.success(`成功添加 ${newAccounts.length} 个公众号，${duplicates.length} 个已存在`)
    } else {
      message.success(`成功添加 ${newAccounts.length} 个公众号`)
    }

    setNewAccountName('')
    setAddAccountModalVisible(false)
  }

  // 监听公众号输入框变化，更新选中状态
  useEffect(() => {
    const accounts = form.getFieldValue('accounts') || ''
    const accountList = accounts.split('\n').filter((a: string) => a.trim())
    setSelectedAccounts(accountList)
  }, [form])

  // 常用网盘链接关键词
  const cloudDriveKeywords = [
    { label: '百度网盘', value: 'https://pan.baidu.com' },
    { label: '阿里云盘', value: 'https://www.aliyundrive.com' },
    { label: '夸克网盘', value: 'https://pan.quark.cn' },
    { label: '天翼云盘', value: 'https://cloud.189.cn' },
    { label: '123云盘', value: 'https://www.123pan.com' },
    { label: '蓝奏云', value: 'https://lanzou' },
    { label: '迅雷云盘', value: 'https://pan.xunlei.com' },
    { label: '微云', value: 'https://share.weiyun.com' },
    { label: '城通网盘', value: 'https://www.ctfile.com' },
    { label: 'OneDrive', value: 'https://onedrive.live.com' },
  ]

  // 处理快捷关键词选择
  const handleQuickKeyword = (value: string) => {
    console.log('选中的网盘关键词:', value)

    // 如果点击的是已选中的项，则取消选择
    if (selectedCloudDrive === value) {
      form.setFieldValue('keywordFilter', '')
      formStore.setKeywordFilter('')
      setSelectedCloudDrive(undefined)
    } else {
      // 直接设置选中的网盘链接关键词
      form.setFieldValue('keywordFilter', value)
      formStore.setKeywordFilter(value)
      setSelectedCloudDrive(value)
    }

    // 手动触发表单变化事件
    handleFormChange()
  }

  // 加载配置
  useEffect(() => {
    const loadConfig = async () => {
      try {
        const cfg = await api.loadConfig()
        console.log('ScrapePage 加载配置:', cfg)
        console.log('ScrapePage 当前 formStore:', {
          includeContent: formStore.includeContent,
          maxPages: formStore.maxPages,
          requestInterval: formStore.requestInterval,
          maxWorkers: formStore.maxWorkers,
        })
        setConfig(cfg)

        // 不要覆盖 formStore，因为用户可能在设置界面修改过
        // formStore 是全局状态，应该保持用户的最新选择
      } catch (error) {
        console.error('Failed to load config:', error)
      }
    }
    if (!config) {
      loadConfig()
    }
  }, [])

  // 从config和formStore初始化表单
  useEffect(() => {
    if (config) {
      form.setFieldsValue({
        accounts: formStore.accounts,
        dateRange: formStore.dateRange,
        maxPages: formStore.maxPages,
        requestInterval: formStore.requestInterval,
        includeContent: formStore.includeContent,
        keywordFilter: formStore.keywordFilter,
      })
    }
  }, [config, formStore.dateRange, formStore.maxPages, formStore.requestInterval, formStore.includeContent])

  // 监听表单变化，保存到store
  const handleFormChange = () => {
    const values = form.getFieldsValue()
    formStore.setAccounts(values.accounts || '')
    formStore.setDateRange(values.dateRange)
    formStore.setMaxPages(Number(values.maxPages) || 10)
    formStore.setRequestInterval(Number(values.requestInterval) || 10)
    formStore.setIncludeContent(values.includeContent)
    formStore.setKeywordFilter(values.keywordFilter || '')
  }

  useEffect(() => {
    // 检查登录状态
    if (!loginStatus?.isLoggedIn) {
      message.warning('请先登录')
      navigate('/login')
      return
    }

    // 定义事件处理函数
    const handleProgress = (prog: any) => {
      console.log('收到进度事件:', prog)

      // 更新全局进度（用于总体进度显示）
      setProgress(prog)

      // 如果进度事件包含账号信息，更新对应账号的进度
      if (prog.message && prog.message.includes('[')) {
        // 从消息中提取账号名称，格式如 "正在获取文章内容 [账号名] (1/10)"
        const match = prog.message.match(/\[([^\]]+)\]/)
        if (match && match[1]) {
          const accountName = match[1]
          updateAccountStatus(accountName, {
            accountName,
            status: 'content',
            message: prog.message.split(']')[1]?.trim() || '正在获取文章内容...',
            articleCount: 0,
            progress: {
              current: prog.current,
              total: prog.total,
            },
          })
        }
      }
    }

    const handleStatus = (status: any) => {
      console.log('收到账号状态:', status)
      console.log('账号名称:', status.accountName, '文章数:', status.articleCount)

      // 检查是否遇到频率限制
      if (status.status === 'error' && status.message &&
          (status.message.includes('频率限制') || status.message.includes('freq control'))) {
        message.warning({
          content: `${status.accountName}: ${status.message}`,
          duration: 5,
        })
        // 设置频率限制标志
        hasFreqLimitErrorRef.current = true
        // 立即取消爬取
        api.cancelScrape().catch(err => console.error('取消爬取失败:', err))
        // 停止爬取状态
        setScrapingInProgress(false)
        setLoading(false)
        return
      }

      // 使用函数式更新，避免闭包问题
      updateAccountStatus(status.accountName, status)

      // 实时更新文章总数 - 追踪每个账号的文章数
      setAccountArticleCounts((prevCounts) => {
        const newCounts = { ...prevCounts, [status.accountName]: status.articleCount }
        console.log('更新后的账号计数:', newCounts)
        // 计算总数 - 显式类型转换
        const counts = Object.values(newCounts) as number[]
        const total = counts.reduce((sum, count) => sum + count, 0)
        console.log('计算的总文章数:', total)
        setTotalArticleCount(total)
        return newCounts
      })
    }

    const handleCompleted = (data: any) => {
      console.log('爬取完成事件:', data)
      console.log('当前 totalArticleCount:', totalArticleCount)
      // 使用实时统计的文章数，而不是事件中的 total
      const actualTotal = totalArticleCount || data.total || 0
      console.log('实际显示的文章数:', actualTotal)

      setScrapingInProgress(false)
      setLoading(false)

      // 如果没有获取到任何文章，不跳转
      if (actualTotal === 0) {
        message.warning('未获取到任何文章')
        return
      }

      message.success(`爬取完成！共获取 ${actualTotal} 篇文章`)

      // 自动添加成功爬取的公众号到常用列表（去重）
      const accounts = form.getFieldValue('accounts') || ''
      const accountList = accounts.split('\n').filter((a: string) => a.trim())

      // 从 localStorage 读取最新的常用列表，避免闭包问题
      const savedFavorites = localStorage.getItem('favoriteAccounts')
      const currentFavorites = savedFavorites ? JSON.parse(savedFavorites) : []

      // 过滤出不在常用列表中的公众号
      const newAccounts = accountList.filter((account: string) => !currentFavorites.includes(account))

      // 如果有新公众号，批量添加
      if (newAccounts.length > 0) {
        const updatedFavorites = [...currentFavorites, ...newAccounts]
        localStorage.setItem('favoriteAccounts', JSON.stringify(updatedFavorites))
        setFavoriteAccounts(updatedFavorites)
        message.success(`已添加 ${newAccounts.length} 个公众号到常用列表`)
      }

      // 不清理状态，保留进度和账号状态供查看
    }

    const handleError = (error: any) => {
      // 调试：打印完整的错误对象
      console.log('收到错误事件:', error)
      console.log('错误类型:', typeof error)
      console.log('错误内容:', JSON.stringify(error))

      // 获取错误信息
      const errorMsg = error?.error || error?.message || String(error)
      console.log('提取的错误信息:', errorMsg)

      // 如果是取消操作，不显示错误
      if (errorMsg.includes('context canceled') || errorMsg.includes('canceled')) {
        console.log('检测到取消操作，不显示错误')
        return
      }

      message.error('爬取失败: ' + errorMsg)
      setScrapingInProgress(false)
      setLoading(false)
    }

    // 设置事件监听，保存取消订阅函数
    const unsubProgress = events.onScrapeProgress(handleProgress)
    const unsubStatus = events.onScrapeStatus(handleStatus)
    const unsubCompleted = events.onScrapeCompleted(handleCompleted)
    const unsubError = events.onScrapeError(handleError)

    // 清理函数
    return () => {
      events.offScrapeProgress(unsubProgress)
      events.offScrapeStatus(unsubStatus)
      events.offScrapeCompleted(unsubCompleted)
      events.offScrapeError(unsubError)
    }
  }, [loginStatus])

  // 开始爬取
  const handleStartScrape = async () => {
    try {
      const values = await form.validateFields()

      // 解析公众号列表
      const accounts = values.accounts
        .split('\n')
        .map((line: string) => line.trim())
        .filter((line: string) => line.length > 0)

      if (accounts.length === 0) {
        message.warning('请输入至少一个公众号名称')
        return
      }

      // 构建配置
      const config: ScrapeConfig = {
        accounts,
        startDate: values.dateRange
          ? values.dateRange[0].format('YYYY-MM-DD')
          : '',
        endDate: values.dateRange
          ? values.dateRange[1].format('YYYY-MM-DD')
          : '',
        recentDays: 0,
        maxPages: Number(values.maxPages) || 10,
        requestInterval: Number(values.requestInterval) || 10,
        includeContent: values.includeContent || false,
        keywordFilter: values.keywordFilter || '',
        maxWorkers: formStore.maxWorkers || 20, // 使用设置页面的并发数配置
      }

      setLoading(true)
      setScrapingInProgress(true)
      // 清理之前的状态，准备新的爬取
      clearAccountStatuses()
      setProgress(null)
      setArticles([])
      setTotalArticleCount(0) // 重置文章计数
      setAccountArticleCounts({}) // 重置账号文章计数
      hasFreqLimitErrorRef.current = false // 重置频率限制标志

      const result = await api.startScrape(config)
      setArticles(result)

      // 爬取完成后，检查是否有文章再决定是否跳转
      setTimeout(() => {
        // 使用 result.length 而不是 totalArticleCount，避免闭包问题
        const articleCount = result?.length || 0
        if (!hasFreqLimitErrorRef.current && articleCount > 0) {
          navigate('/results')
        } else if (articleCount === 0) {
          message.warning('未获取到任何文章，请检查配置')
        }
      }, 1500)
    } catch (error: any) {
      // 如果是取消操作，不显示错误
      const errorMsg = error?.message || error?.toString() || '未知错误'
      if (errorMsg.includes('context canceled') || errorMsg.includes('canceled')) {
        console.log('爬取被取消')
        return
      }
      message.error('爬取失败: ' + errorMsg)
      setScrapingInProgress(false)
      setLoading(false)
    }
  }

  // 取消爬取
  const handleCancelScrape = async () => {
    try {
      await api.cancelScrape()
      message.info('已取消爬取')
      setScrapingInProgress(false)
      setLoading(false)
      setProgress(null) // 立即清除进度条

      // 更新所有账号状态为已取消
      accountStatuses.forEach(status => {
        if (status.status !== 'completed' && status.status !== 'error') {
          updateAccountStatus(status.accountName, {
            ...status,
            status: 'error',
            message: '已取消'
          })
        }
      })

      // 不跳转到结果页面，不清除账号状态
      // 账号状态容器保持显示
    } catch (error: any) {
      // 取消操作不应该失败，如果失败也只是提示
      message.info('取消操作: ' + (error.message || '未知错误'))
      setScrapingInProgress(false)
      setLoading(false)
      setProgress(null)

      // 更新所有账号状态为已取消
      accountStatuses.forEach(status => {
        if (status.status !== 'completed' && status.status !== 'error') {
          updateAccountStatus(status.accountName, {
            ...status,
            status: 'error',
            message: '已取消'
          })
        }
      })
    }
  }

  // 状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'searching':
      case 'fetching':
      case 'content':
        return <LoadingOutlined style={{ color: '#1890ff' }} />
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />
      case 'error':
        return <CloseCircleOutlined style={{ color: '#ff4d4f' }} />
      default:
        return null
    }
  }

  // 表格列定义
  const columns = [
    {
      title: '公众号',
      dataIndex: 'accountName',
      key: 'accountName',
      width: 120,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 280,
      render: (status: string, record: AccountStatus) => (
        <div style={{ width: '100%' }}>
          {/* 如果有进度条，只显示进度条 */}
          {record.progress && record.progress.total > 0 ? (
            <Progress
              percent={Math.round((record.progress.current / record.progress.total) * 100)}
              size="small"
              status={status === 'completed' ? 'success' : status === 'error' ? 'exception' : 'active'}
              format={(percent) => `${record.progress?.current || 0}/${record.progress?.total || 0}`}
              style={{ marginBottom: 0 }}
            />
          ) : (
            /* 没有进度条时显示状态文字 */
            <div style={{ display: 'flex', alignItems: 'center' }}>
              {getStatusIcon(status)}
              <span style={{ marginLeft: 8 }}>{record.message}</span>
            </div>
          )}
        </div>
      ),
    },
    {
      title: '文章数',
      dataIndex: 'articleCount',
      key: 'articleCount',
      width: 80,
      align: 'center' as const,
    },
  ]

  return (
    <div style={{ height: '100%', overflow: 'hidden' }}>
      <Space direction="vertical" size="small" style={{ width: '100%', height: '100%' }}>
        {/* 配置表单 */}
        <Card title="爬取配置" styles={{ body: { padding: 16 } }} style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.3)' }}>
          <Form
            form={form}
            onValuesChange={handleFormChange}
            size="small"
          >
            {/* 第一行：公众号 */}
            <div style={{ display: 'flex', gap: '8px', marginBottom: 8 }}>
              <Tooltip title="每行一个公众号名称" placement="topLeft">
                <Form.Item
                  label="公众号列表"
                  name="accounts"
                  rules={[{ required: true, message: '' }]}
                  labelCol={{ flex: '90px' }}
                  style={{ marginBottom: 0, flex: 1 }}
                >
                  <TextArea
                    rows={2}
                    placeholder="每行一个公众号名称，例如：人民日报、新华社"
                    disabled={isScrapingInProgress}
                    style={{ resize: 'none', height: 30 }}
                    onChange={(e) => {
                      const accountList = e.target.value.split('\n').filter((a: string) => a.trim())
                      setSelectedAccounts(accountList)
                    }}
                  />
                </Form.Item>
              </Tooltip>

              <Dropdown
                menu={{
                  items: [
                    ...favoriteAccounts.map((account) => ({
                      key: account,
                      label: (
                        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', minWidth: 180 }}>
                          <span>{account}</span>
                          <Space size={4}>
                            {selectedAccounts.includes(account) && (
                              <CheckCircleOutlined style={{ color: '#52c41a' }} />
                            )}
                            <Button
                              type="text"
                              size="small"
                              danger
                              icon={<CloseCircleOutlined />}
                              onClick={(e) => {
                                e.stopPropagation()
                                removeFromFavorites(account)
                              }}
                            />
                          </Space>
                        </div>
                      ),
                      onClick: () => toggleAccountSelection(account),
                    })),
                    { type: 'divider' },
                    {
                      key: 'add',
                      label: (
                        <div style={{ color: '#07C160' }}>
                          <PlusOutlined /> 添加常用公众号
                        </div>
                      ),
                      onClick: () => setAddAccountModalVisible(true),
                    },
                  ],
                }}
                trigger={['click']}
                disabled={isScrapingInProgress}
              >
                <Button
                  icon={<StarOutlined />}
                  disabled={isScrapingInProgress}
                  style={{
                    height: 30,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    padding: '4px 15px'
                  }}
                >
                  常用
                </Button>
              </Dropdown>
            </div>

            {/* 添加常用公众号弹窗 */}
            <Modal
              title="添加常用公众号"
              open={addAccountModalVisible}
              onOk={handleAddNewAccount}
              onCancel={() => {
                setAddAccountModalVisible(false)
                setNewAccountName('')
              }}
              okText="添加"
              cancelText="取消"
            >
              <TextArea
                rows={5}
                placeholder="每行一个公众号名称，支持批量添加&#10;例如：&#10;人民日报&#10;新华社&#10;央视新闻"
                value={newAccountName}
                onChange={(e) => setNewAccountName(e.target.value)}
                style={{ resize: 'none' }}
              />
            </Modal>

            {/* 第二行：日期范围和关键词 */}
            <div style={{ display: 'flex', gap: '16px', marginBottom: 8 }}>
              <Form.Item
                label="日期范围"
                name="dateRange"
                labelCol={{ flex: '90px' }}
                style={{ marginBottom: 0, flex: 1 }}
              >
                <RangePicker
                  style={{ width: '100%' }}
                  disabled={isScrapingInProgress ? [true, true] : [false, true]}
                  allowEmpty={[false, false]}
                />
              </Form.Item>

              <div style={{ flex: 1, display: 'flex', alignItems: 'center' }}>
                <span style={{ width: '90px', textAlign: 'right', marginRight: '8px' }}>关键词</span>
                <div style={{ flex: 1, display: 'flex', gap: '8px', alignItems: 'center' }}>
                  <Form.Item
                    name="keywordFilter"
                    style={{ marginBottom: 0, flex: 1, minWidth: 0 }}
                  >
                    <Input
                      placeholder="可选，筛选标题"
                      disabled={isScrapingInProgress}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                  <Select
                    placeholder="网盘"
                    style={{ width: 120, flexShrink: 0 }}
                    disabled={isScrapingInProgress}
                    onChange={handleQuickKeyword}
                    value={selectedCloudDrive}
                    allowClear
                    onClear={() => {
                      form.setFieldValue('keywordFilter', '')
                      formStore.setKeywordFilter('')
                      setSelectedCloudDrive(undefined)
                      handleFormChange()
                    }}
                    options={cloudDriveKeywords}
                    popupMatchSelectWidth={false}
                    classNames={{
                      popup: {
                        root: 'cloud-drive-select-dropdown'
                      }
                    }}
                  />
                </div>
              </div>
            </div>

            {/* 第三行：数字配置和开关 */}
            <div style={{ display: 'flex', gap: '86px', marginBottom: 8 }}>
              <Form.Item
                label="最大页数"
                name="maxPages"
                labelCol={{ flex: '90px' }}
                style={{ marginBottom: 0, flex: '0 0 200px' }}
              >
                <Input
                  type="number"
                  min={1}
                  max={100}
                  style={{ width: '100%' }}
                  disabled={isScrapingInProgress}
                  suffix="页"
                />
              </Form.Item>

              <Form.Item
                label="请求间隔"
                name="requestInterval"
                labelCol={{ flex: '90px' }}
                style={{ marginBottom: 0, flex: '0 0 200px', marginRight: '-90px' }}
              >
                <Input
                  type="number"
                  min={1}
                  max={60}
                  style={{ width: '100%' }}
                  disabled={isScrapingInProgress}
                  suffix="秒"
                />
              </Form.Item>

              <div style={{ flex: 1, display: 'flex', alignItems: 'center' }}>
                <span style={{ width: '125px', textAlign: 'right', marginRight: '8px' }}>获取正文</span>
                <Form.Item
                  name="includeContent"
                  valuePropName="checked"
                  style={{ marginBottom: 0 }}
                >
                  <Switch disabled={isScrapingInProgress} />
                </Form.Item>
              </div>
            </div>

            {/* 按钮行 */}
            <div style={{ paddingTop: 4 }}>
              {!isScrapingInProgress ? (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={handleStartScrape}
                  loading={loading}
                  block
                >
                  开始爬取
                </Button>
              ) : (
                <Button
                  danger
                  icon={<StopOutlined />}
                  onClick={handleCancelScrape}
                  block
                >
                  取消爬取
                </Button>
              )}
            </div>
          </Form>
        </Card>

        {/* 账号状态 - 始终显示容器 */}
        <Card
          title={
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span>账号状态</span>
              {totalArticleCount > 0 && (
                <span style={{ fontSize: 12, color: '#999', fontWeight: 'normal' }}>
                  已获取 {totalArticleCount} 篇文章
                </span>
              )}
            </div>
          }
          size="small"
          styles={{ body: { padding: 16 } }}
          style={{ boxShadow: '0 2px 8px rgba(0,0,0,0.3)', flex: 1, overflow: 'hidden' }}
        >
          {accountStatuses.length > 0 ? (
            <Table
              dataSource={accountStatuses.map((item, index) => ({ ...item, key: `${item.accountName}-${index}` }))}
              columns={columns}
              pagination={false}
              size="small"
              scroll={{ y: 300 }}
            />
          ) : (
            <div style={{ textAlign: 'center', color: '#999', padding: '40px 0' }}>
              暂无爬取任务
            </div>
          )}
        </Card>

        {/* 结果提示 */}
        {articles && articles.length > 0 && !isScrapingInProgress && (
          <Alert
            message={`已获取 ${articles.length} 篇文章`}
            type="success"
            showIcon
            action={
              <Button size="small" onClick={() => navigate('/results')}>
                查看结果
              </Button>
            }
          />
        )}
      </Space>
    </div>
  )
}

export default ScrapePage
