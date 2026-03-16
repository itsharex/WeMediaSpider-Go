import React, { useEffect, useRef } from 'react'
import * as echarts from 'echarts'
import 'echarts-wordcloud'

interface WordCloudChartProps {
  data: Array<{ word: string; count: number }>
}

const WordCloudChart: React.FC<WordCloudChartProps> = ({ data }) => {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<echarts.ECharts | null>(null)

  useEffect(() => {
    if (!chartRef.current || !data || data.length === 0) return

    // 初始化图表
    if (!chartInstance.current) {
      chartInstance.current = echarts.init(chartRef.current)
    }

    // 转换数据格式
    const wordCloudData = data.map(item => ({
      name: item.word,
      value: item.count,
    }))

    // 配置词云
    const option = {
      tooltip: {
        show: true,
        formatter: (params: any) => {
          return `${params.name}: ${params.value}`
        },
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        borderColor: '#07C160',
        borderWidth: 1,
        textStyle: {
          color: '#fff',
        },
      },
      series: [
        {
          type: 'wordCloud',
          shape: 'circle',
          // 词云大小范围
          sizeRange: [12, 60],
          // 旋转角度范围
          rotationRange: [-45, 45],
          rotationStep: 45,
          // 词语间距
          gridSize: 8,
          // 绘制超出边界
          drawOutOfBound: false,
          // 布局动画
          layoutAnimation: true,
          // 文字样式
          textStyle: {
            fontFamily: 'sans-serif',
            fontWeight: 'bold',
            // 颜色函数 - 绿色系渐变
            color: function () {
              const colors = [
                '#07C160',
                '#52c41a',
                '#73d13d',
                '#95de64',
                '#b7eb8f',
                '#d9f7be',
                '#1890ff',
                '#40a9ff',
                '#69c0ff',
              ]
              return colors[Math.floor(Math.random() * colors.length)]
            },
          },
          emphasis: {
            focus: 'self',
            textStyle: {
              textShadowBlur: 10,
              textShadowColor: '#07C160',
            },
          },
          data: wordCloudData,
        },
      ],
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

export default WordCloudChart
