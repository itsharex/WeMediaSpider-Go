package timeutil

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

const (
	// 中国国家授时中心 NTP 服务器
	ChinaNTPServer = "ntp.ntsc.ac.cn"
	// 时间同步间隔
	SyncInterval = 1 * time.Hour
)

// 备用 NTP 服务器列表
var ntpServers = []string{
	"ntp.ntsc.ac.cn",      // 中国国家授时中心
	"ntp.aliyun.com",      // 阿里云
	"ntp1.aliyun.com",     // 阿里云备用1
	"ntp2.aliyun.com",     // 阿里云备用2
	"time.windows.com",    // Windows 时间服务器
	"time.apple.com",      // Apple 时间服务器
	"pool.ntp.org",        // NTP Pool
}

var (
	chinaLocation *time.Location
	timeOffset    time.Duration
	lastSync      time.Time
	mu            sync.RWMutex
	syncOnce      sync.Once
)

func init() {
	// 加载中国时区 (UTC+8)
	var err error
	chinaLocation, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 如果加载失败，使用固定偏移
		chinaLocation = time.FixedZone("CST", 8*3600)
	}
}

// Now 返回中国时间（东八区，从 NTP 服务器同步）
func Now() time.Time {
	mu.RLock()
	offset := timeOffset
	mu.RUnlock()

	// 获取 UTC 时间并加上偏移，然后转换为中国时区
	utcTime := time.Now().UTC().Add(offset)
	return utcTime.In(chinaLocation)
}

// SyncTime 从 NTP 服务器同步时间
func SyncTime() error {
	var lastErr error

	// 尝试所有 NTP 服务器
	for _, server := range ntpServers {
		offset, err := getTimeOffset(server)
		if err != nil {
			lastErr = err
			continue
		}

		// 同步成功
		mu.Lock()
		timeOffset = offset
		lastSync = time.Now().UTC()
		mu.Unlock()

		return nil
	}

	// 所有服务器都失败
	if lastErr != nil {
		return fmt.Errorf("failed to sync time from all NTP servers: %w", lastErr)
	}

	return fmt.Errorf("no NTP servers available")
}

// getTimeOffset 获取与 NTP 服务器的时间偏移
func getTimeOffset(server string) (time.Duration, error) {
	// 设置超时时间为 3 秒
	response, err := ntp.QueryWithOptions(server, ntp.QueryOptions{
		Timeout: 3 * time.Second,
	})
	if err != nil {
		return 0, err
	}

	// NTP 返回的是相对于本地时间的偏移
	// 我们需要计算相对于 UTC 的偏移
	return response.ClockOffset, nil
}

// StartAutoSync 启动自动时间同步
func StartAutoSync() {
	syncOnce.Do(func() {
		// 立即同步一次
		if err := SyncTime(); err != nil {
			// 如果同步失败，使用系统时间
			fmt.Printf("Initial NTP sync failed: %v, using system time\n", err)
		}

		// 定期同步
		go func() {
			ticker := time.NewTicker(SyncInterval)
			defer ticker.Stop()

			for range ticker.C {
				if err := SyncTime(); err != nil {
					fmt.Printf("NTP sync failed: %v\n", err)
				}
			}
		}()
	})
}

// GetLastSyncTime 获取上次同步时间（中国时区）
func GetLastSyncTime() time.Time {
	mu.RLock()
	defer mu.RUnlock()
	if lastSync.IsZero() {
		return time.Time{}
	}
	return lastSync.In(chinaLocation)
}

// GetTimeOffset 获取当前时间偏移
func GetTimeOffset() time.Duration {
	mu.RLock()
	defer mu.RUnlock()
	return timeOffset
}

// GetChinaLocation 获取中国时区
func GetChinaLocation() *time.Location {
	return chinaLocation
}
