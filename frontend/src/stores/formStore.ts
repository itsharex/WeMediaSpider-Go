import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import dayjs, { Dayjs } from 'dayjs'
import { GetTimeInfo } from '../../wailsjs/go/app/App'

interface FormState {
  accounts: string
  dateRange: [Dayjs, Dayjs] | null
  maxPages: number
  requestInterval: number
  maxWorkers: number
  includeContent: boolean
  keywordFilter: string
  setAccounts: (accounts: string) => void
  setDateRange: (dateRange: [Dayjs, Dayjs] | null) => void
  setMaxPages: (maxPages: number) => void
  setRequestInterval: (requestInterval: number) => void
  setMaxWorkers: (maxWorkers: number) => void
  setIncludeContent: (includeContent: boolean) => void
  setKeywordFilter: (keywordFilter: string) => void
  reset: () => void
  initChinaTime: () => Promise<void>
}

// 获取中国时间的今天
const getChinaToday = async (): Promise<Dayjs> => {
  try {
    const timeInfo = await GetTimeInfo()
    // 使用 currentDate 字段，避免时区转换问题
    return dayjs(timeInfo.currentDate)
  } catch (error) {
    console.error('获取中国时间失败，使用本地时间:', error)
    return dayjs()
  }
}

const defaultState = {
  accounts: '',
  dateRange: [dayjs().subtract(30, 'day'), dayjs()] as [Dayjs, Dayjs],
  maxPages: 10,
  requestInterval: 10,
  maxWorkers: 20,
  includeContent: false,
  keywordFilter: '',
}

export const useFormStore = create<FormState>()(
  persist(
    (set) => ({
      ...defaultState,
      setAccounts: (accounts) => set({ accounts }),
      setDateRange: (dateRange) => set((state) => {
        if (dateRange) {
          // endDate 始终锁定为当天
          const currentEnd = state.dateRange ? state.dateRange[1] : dayjs()
          return { dateRange: [dateRange[0], currentEnd] }
        }
        return { dateRange }
      }),
      setMaxPages: (maxPages) => set({ maxPages }),
      setRequestInterval: (requestInterval) => set({ requestInterval }),
      setMaxWorkers: (maxWorkers) => set({ maxWorkers }),
      setIncludeContent: (includeContent) => set({ includeContent }),
      setKeywordFilter: (keywordFilter) => set({ keywordFilter }),
      reset: () => set(defaultState),
      initChinaTime: async () => {
        const chinaToday = await getChinaToday()
        set({ dateRange: [chinaToday.subtract(30, 'day'), chinaToday] })
      },
    }),
    {
      name: 'form-storage',
      // 自定义序列化，处理 dayjs 对象
      storage: {
        getItem: (name) => {
          const str = localStorage.getItem(name)
          if (!str) return null
          const { state } = JSON.parse(str)
          // 恢复 dayjs 对象
          if (state.dateRange) {
            state.dateRange = [
              dayjs(state.dateRange[0]),
              dayjs(state.dateRange[1])
            ]
          }
          // 确保 accounts 始终为空
          state.accounts = ''
          return { state }
        },
        setItem: (name, value) => {
          const str = JSON.stringify(value)
          localStorage.setItem(name, str)
        },
        removeItem: (name) => localStorage.removeItem(name),
      },
      // 排除 accounts 字段，不持久化公众号列表
      partialize: (state: FormState) => ({
        dateRange: state.dateRange,
        maxPages: state.maxPages,
        requestInterval: state.requestInterval,
        maxWorkers: state.maxWorkers,
        includeContent: state.includeContent,
        keywordFilter: state.keywordFilter,
      }),
    }
  )
)
