package tool

import (
	"strings"
	"time"
)

// Date 单例Date并导出
var Date = &date{}

type date struct {
}

// Today 当天开始时间
func (s *date) Today() time.Time {
	t, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	return t
}

// MonthStart 本月开始
func (s *date) MonthStart() time.Time {
	now := time.Now()
	last := now.AddDate(0, -1, -now.Day()+1)
	return time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, last.Location())
}

// MonthEnd 本月结束
func (s *date) MonthEnd() time.Time {
	now := time.Now()
	last := now.AddDate(0, 1, -now.Day())
	return time.Date(last.Year(), last.Month(), last.Day(), 23, 59, 59, 0, last.Location())
}

// WeekEnd 本周开始, 作为查询，应该大于等于此值
func (s *date) WeekStart() time.Time {
	now := time.Now()
	//改为周一为 0
	week := 7
	if int(now.Weekday()) > 0 {
		week = int(now.Weekday()) - 1
	}
	last := now.AddDate(0, 0, -week)
	return time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, last.Location())
}

// WeekEnd 本周结束, 作为查询，应该小于此值
func (s *date) WeekEnd() time.Time {
	now := time.Now()
	//改为周一为 0
	week := 7
	if int(now.Weekday()) > 0 {
		week = int(now.Weekday()) - 1
	}
	last := now.AddDate(0, 0, 7-week)
	return time.Date(last.Year(), last.Month(), last.Day(), 0, 0, 0, 0, last.Location())
}

// 获取自然周的开始时间，从周一开始
func (s *date) GetWeekStartTime(timestamp int64) time.Time {
	now := time.Now()
	if timestamp > 0 {
		now = time.Unix(timestamp, 0)
	}
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return weekStartDate
}

// 传入的日期符合格式2021-09-20, 使用Local时区，例如中国为+0800
func (s *date) Parse(t string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", t, time.Local)
}

// GetDateTime 根据秒时间戳获取年月日时分秒, 第二个参数false不带时分秒
func (s *date) GetDateTime(sec int64, dateType bool) string {
	var template string
	if dateType {
		template = "2006-01-02 15:04:05"
	} else {
		template = "2006-01-02"
	}
	return time.Unix(sec, 0).Local().Format(template)
}

// GetUnix 根据日期字符串,获取时间戳,example[2021-11-05或者2021-11-05 23:59:59]
func (s *date) GetUnix(t string) int64 {
	//参数空,为当前时间戳
	if t == "" {
		return time.Now().Unix()
	}
	sArr := strings.Split(t, " ")
	tmp := "2006-01-02"
	if len(sArr) == 2 {
		tmp = "2006-01-02 15:04:05"
	}
	value, err := time.ParseInLocation(tmp, t, time.Local)
	if err != nil {
		return 0
	}
	return value.Unix()
}

// 现在之前的最近一个周X的开始时间
func (s *date) NearestWeekday(day time.Weekday) time.Time {
	now := time.Now()
	diff := (now.Weekday() - day) % 7
	if diff > 0 {
		now = now.AddDate(0, 0, -int(diff))
	}
	return s.DateByTime(now)
}

func (s *date) DateByTime(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
func (s *date) TodayEnd() time.Time {
	t := time.Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, t.Location())
}

// LastWeekStart 上周开始
func (s *date) LastWeekStart() time.Time {
	t := s.GetWeekStartTime(0)
	t = t.AddDate(0, 0, -7)
	return t
}

func (s *date) Age(timestamp int64) int64 {
	now := time.Now().Unix()
	return (now - int64(timestamp)) / (86400 * 365)
}

func (s *date) InPeriod(start string, end string) bool {
	now := time.Now().Unix()

	return now < s.GetUnix(end) && now >= s.GetUnix(start)
}

// 将剩余秒数转换成倒计时天、小时、分
func (s *date) ResolveTime(seconds int64) (day int64, hour int64, minute int64) {
	if seconds > 86400 {
		day = seconds / 86400
	}
	leftSecs := seconds - day*86400
	if leftSecs > 3600 {
		hour = leftSecs / 3600
	}
	leftSecs -= hour * 3600
	if leftSecs > 60 {
		minute = leftSecs / 60
	}
	return
}
