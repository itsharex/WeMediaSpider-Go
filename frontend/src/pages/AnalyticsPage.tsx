import React, { useState, useEffect } from 'react'
import { Card, DatePicker, Button, Spin, App, Select } from 'antd'
import { ReloadOutlined } from '@ant-design/icons'
import dayjs, { Dayjs } from 'dayjs'
import { GetAnalyticsData, ClearAnalyticsCache, GetAllAccountNames, GetTimeInfo } from '../../wailsjs/go/app/App'
import TimeDistributionChart from '../components/charts/TimeDistributionChart'
import WordCloudChart from '../components/charts/WordCloudChart'

const { RangePicker } = DatePicker

interface AccountTimeDistribution {
  accountName: string
  data: Array<{ date: string; count: number }>
}

interface AnalyticsData {
  timeDistribution: AccountTimeDistribution[]
  topKeywords: Array<{ word: string; count: number }>
  cachedAt: string
}

const AnalyticsPage: React.FC = () => {
  const { message: antMessage } = App.useApp()
  const [loading, setLoading] = useState(false)
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [dateRange, setDateRange] = useState<[Dayjs, Dayjs]>([
    dayjs().subtract(30, 'days'),
    dayjs(),
  ])
  const [allAccounts, setAllAccounts] = useState<string[]>([])
  const [selectedAccounts, setSelectedAccounts] = useState<string[]>([])

  // 初始化中国时间并加载数据
  useEffect(() => {
    const initChinaTime = async () => {
      try {
        const timeInfo = await GetTimeInfo()
        // 使用 currentDate 字段，避免时区转换问题
        const chinaToday = dayjs(timeInfo.currentDate)
        const newDateRange: [Dayjs, Dayjs] = [chinaToday.subtract(30, 'days'), chinaToday]
        setDateRange(newDateRange)
        console.log('初始化日期范围:', {
          start: newDateRange[0].format('YYYY-MM-DD'),
          end: newDateRange[1].format('YYYY-MM-DD'),
          currentDate: timeInfo.currentDate
        })

        // 初始化完成后，清除缓存并加载数据
        await ClearAnalyticsCache()
        // 使用新的日期范围加载数据
        const startDate = newDateRange[0].format('YYYY-MM-DD')
        const endDate = newDateRange[1].format('YYYY-MM-DD')
        const result = await GetAnalyticsData(startDate, endDate, selectedAccounts, true)
        setData(result)
      } catch (error) {
        console.error('初始化失败:', error)
      } finally {
        setLoading(false)
      }
    }

    setLoading(true)
    initChinaTime()
  }, [])

  // 加载公众号列表
  useEffect(() => {
    const loadAccounts = async () => {
      try {
        const accounts = await GetAllAccountNames()
        setAllAccounts(accounts || [])
      } catch (error) {
        console.error('加载公众号列表失败:', error)
      }
    }
    loadAccounts()
  }, [])

  // 加载分析数据
  const loadAnalyticsData = async (forceRefresh = false) => {
    try {
      setLoading(true)
      const startDate = dateRange[0].format('YYYY-MM-DD')
      const endDate = dateRange[1].format('YYYY-MM-DD')

      const result = await GetAnalyticsData(startDate, endDate, selectedAccounts, forceRefresh)
      setData(result)

      if (forceRefresh) {
        antMessage.success('数据已刷新')
      }
    } catch (error) {
      console.error('加载分析数据失败:', error)
      antMessage.error('加载分析数据失败')
    } finally {
      setLoading(false)
    }
  }

  // 强制刷新
  const handleRefresh = async () => {
    await ClearAnalyticsCache()
    await loadAnalyticsData(true)
  }

  // 日期范围变化
  const handleDateRangeChange = (dates: any) => {
    if (dates && dates[0]) {
      // endDate 锁定为当天
      setDateRange([dates[0], dateRange[1]])
    }
  }

  // 应用日期筛选 - 强制刷新，不使用缓存
  const handleApplyFilter = () => {
    loadAnalyticsData(true)
  }

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      height: '100%',
      overflow: 'hidden',
      padding: '12px 24px'
    }}>
      {/* 顶部控制栏 */}
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        marginBottom: 16,
        height: 60
      }}>
        <RangePicker
          value={dateRange}
          onChange={handleDateRangeChange}
          format="YYYY-MM-DD"
          style={{ width: 280 }}
          disabled={[false, true]}
        />
        <Select
          mode="multiple"
          placeholder="选择公众号（不选则全部）"
          value={selectedAccounts}
          onChange={setSelectedAccounts}
          style={{ width: 300 }}
          maxTagCount="responsive"
        >
          {allAccounts.map(name => (
            <Select.Option key={name} value={name}>
              {name}
            </Select.Option>
          ))}
        </Select>
        <Button type="primary" onClick={handleApplyFilter}>
          应用筛选
        </Button>
        <Button icon={<ReloadOutlined />} onClick={handleRefresh} loading={loading}>
          刷新
        </Button>
        {data?.cachedAt && (
          <span style={{ color: 'rgba(255, 255, 255, 0.45)', fontSize: 12, marginLeft: 'auto' }}>
            缓存时间: {data.cachedAt}
          </span>
        )}
      </div>

      {/* 图表区域 */}
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
        flex: 1,
        minHeight: 0
      }}>
        {loading ? (
          <div style={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center'
          }}>
            <Spin size="large" tip="加载分析数据中..." />
          </div>
        ) : (
          <>
            {/* 时间趋势图 */}
            <Card
              title="文章发布趋势"
              style={{
                background: 'rgba(255, 255, 255, 0.05)',
                flex: 1,
                overflow: 'hidden'
              }}
              bodyStyle={{ height: 'calc(100% - 40px)', padding: 16 }}
            >
              {data?.timeDistribution && data.timeDistribution.length > 0 ? (
                <TimeDistributionChart data={data.timeDistribution} />
              ) : (
                <div style={{ color: 'rgba(255, 255, 255, 0.45)', textAlign: 'center', paddingTop: 80 }}>
                  暂无数据
                </div>
              )}
            </Card>

            {/* 关键词词云 */}
            <Card
              title="热门关键词"
              style={{
                background: 'rgba(255, 255, 255, 0.05)',
                flex: 1,
                overflow: 'hidden'
              }}
              bodyStyle={{ height: 'calc(100% - 40px)', padding: 16 }}
            >
              {data?.topKeywords && data.topKeywords.length > 0 ? (
                <WordCloudChart data={data.topKeywords} />
              ) : (
                <div style={{ color: 'rgba(255, 255, 255, 0.45)', textAlign: 'center', paddingTop: 80 }}>
                  暂无数据
                </div>
              )}
            </Card>
          </>
        )}
      </div>
    </div>
  )
}

export default AnalyticsPage
