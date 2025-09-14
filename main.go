package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ENDPOINT string = "http://srv.msk01.gigacorp.local"

type Metric struct {
	LoadCPUAverage int64
	AllRAM         int64
	LoadRAM        int64
	SpaceDisk      int64
	UseDisk        int64
	BandwidthBps   int64
	LoadBps        int64
}

func GetData() (string, error) {
	resp, err := http.Get(ENDPOINT)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("HTTP-запрос завершился с ошибкой: " + resp.Status)
	}

	return string(body), nil
}

func ParseAnswer(data string) (*Metric, error) {
	parts := strings.Split(data, ",")

	if len(parts) != 7 {
		return nil, fmt.Errorf("получено %d значений", len(parts))
	}

	var metric Metric
	var err error

	metric.LoadCPUAverage, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга LoadCPUAverage: %v", err)
	}
	metric.AllRAM, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга AllRAM: %v", err)
	}
	metric.LoadRAM, err = strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга LoadRAM: %v", err)
	}
	metric.SpaceDisk, err = strconv.ParseInt(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга SpaceDisk: %v", err)
	}
	metric.UseDisk, err = strconv.ParseInt(parts[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга UseDisk: %v", err)
	}
	metric.BandwidthBps, err = strconv.ParseInt(parts[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга BandwidthBps: %v", err)
	}
	metric.LoadBps, err = strconv.ParseInt(parts[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга LoadBps: %v", err)
	}

	return &metric, nil
}

func CheckMetric(m *Metric) string {
	var out string
	if m.LoadCPUAverage >= 30 {
		out += fmt.Sprintf("Load Average is too high: %d\n", m.LoadCPUAverage)
	}

	percentage := int(float32(m.LoadRAM) / float32(m.AllRAM) * 100)
	if percentage >= 80 {
		out += fmt.Sprintf("Memory usage too high: %d%%\n", percentage)
	}

	percentage = int(float32(m.UseDisk) / float32(m.SpaceDisk) * 100)
	if percentage >= 90 {
		out += fmt.Sprintf("Free disk space is too low: %d Mb left\n", (m.SpaceDisk-m.UseDisk)/1024/1024)
	}

	percentage = int(float32(m.LoadBps) / float32(m.BandwidthBps) * 100)
	if percentage >= 90 {
		out += fmt.Sprintf("Network bandwidth usage high: %d Mbit/s available\n", (m.BandwidthBps-m.LoadBps)/1_000_000)
	}

	return out
}
func main() {
	count := 0
	for {
		time.Sleep(time.Millisecond * 500)
		if count == 3 {
			fmt.Println("Unable to fetch server statistic")
			count = 0
		}

		answer, err := GetData()
		if err != nil {
			count += 1
			continue
		}

		m, err := ParseAnswer(answer)

		if err != nil {
			count += 1
			continue
		}

		s := CheckMetric(m)

		if s != "" {
			fmt.Print(s)
			count = 0
			continue
		}
	}
}
