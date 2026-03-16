import React, { useEffect, useRef } from 'react'
import * as echarts from 'echarts'

interface AccountTimeDistribution {
  accountName: string
  data: Array<{ date: string; count: number }>
}

interface TimeDistributionChartProps {
  data: AccountTimeDistribution[]
}

const TimeDistributionChart: React.FC<TimeDistributionChartProps> = ({ data }) => {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)

  useEffect(() => {
    if (!chartRef.current || !data || data.length === 0) return

    // 初始化图表
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current)
    }

    // 收集所有日期（去重并排序）
    const allDatesSet = new Set<string>()
    data.forEach(account => {
      account.data.forEach(item => allDatesSet.add(item.date))
    })

    if (allDatesSet.size === 0) {
      console.warn('没有可用的日期数据')
      return
    }

    const sortedDates = Array.from(allDatesSet).sort()

    // 生成完整的日期范围（从最早到最晚，填充缺失日期）
    const allDates: string[] = []
    const startDate = new Date(sortedDates[0])
    const endDate = new Date(sortedDates[sortedDates.length - 1])

    console.log('图表日期范围:', {
      start: sortedDates[0],
      end: sortedDates[sortedDates.length - 1],
      totalDays: Math.ceil((endDate.getTime() - startDate.getTime()) / (1000 * 60 * 60 * 24)) + 1
    })

    const currentDate = new Date(startDate)
    while (currentDate <= endDate) {
      const dateStr = currentDate.toISOString().split('T')[0]
      allDates.push(dateStr)
      currentDate.setDate(currentDate.getDate() + 1)
    }

    console.log('生成的完整日期数组长度:', allDates.length)

    // 为每个公众号生成一条线
    const colors = [
      '#07C160', '#1890ff', '#faad14', '#f5222d', '#722ed1',
      '#13c2c2', '#52c41a', '#eb2f96', '#fa8c16', '#2f54eb'
    ]

    const series = data.map((account, index) => {
      // 创建日期到数量的映射
      const dateMap = new Map<string, number>()
      account.data.forEach(item => {
        dateMap.set(item.date, item.count)
      })

      // 填充所有日期的数据（缺失的为0）
      const counts = allDates.map(date => dateMap.get(date) || 0)

      console.log(`${account.accountName}: 原始数据点=${account.data.length}, 填充后=${counts.length}`)

      const color = colors[index % colors.length]

      return {
        name: account.accountName,
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        sampling: 'lttb',
        itemStyle: {
          color: color,
        },
        lineStyle: {
          width: 2,
          color: color,
        },
        emphasis: {
          focus: 'series',
        },
        data: counts,
      }
    })

    // 配置多线条图表
    const option = {
      tooltip: {
        trigger: 'axis',
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        borderColor: '#07C160',
        borderWidth: 1,
        textStyle: {
          color: '#fff',
        },
        axisPointer: {
          type: 'cross',
          label: {
            backgroundColor: '#07C160',
          },
        },
      },
      legend: {
        data: data.map(account => account.accountName),
        textStyle: {
          color: 'rgba(255, 255, 255, 0.65)',
          fontSize: 11,
        },
        top: 0,
        type: 'scroll',
      },
      grid: {
        left: '3%',
        right: '5%',
        bottom: '8%',
        top: '15%',
        containLabel: true,
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: allDates,
        axisLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.2)',
          },
        },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.65)',
          fontSize: 11,
          rotate: allDates.length > 30 ? 45 : 0, // 日期多时旋转标签
          interval: 'auto', // 自动计算显示间隔
        },
        axisTick: {
          alignWithLabel: true,
        },
      },
      yAxis: {
        type: 'value',
        axisLine: {
          show: false,
        },
        axisTick: {
          show: false,
        },
        axisLabel: {
          color: 'rgba(255, 255, 255, 0.65)',
          fontSize: 11,
        },
        splitLine: {
          lineStyle: {
            color: 'rgba(255, 255, 255, 0.1)',
          },
        },
      },
      series: series,
    }

    chartInstance.current.setOption(option)

    // 监听容器大小变化
    const resizeObserver = new ResizeObserver(() => {
      chartInstance.current?.resize()
    })

    resizeObserver.observe(chartRef.current)

    return () => {
      resizeObserver.disconnect()
    }
  }, [data])

  // 清理
  useEffect(() => {
    return () => {
      chartInstance.current?.dispose()
      chartInstance.current = null
    }
  }, [])

  return (
    <div
      ref={chartRef}
      style={{
        width: '100%',
        height: '100%',
        minHeight: 200,
      }}
    />
  )
}

export default TimeDistributionChart
