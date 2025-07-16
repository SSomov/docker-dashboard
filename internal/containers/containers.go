package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

type Container struct {
	ID             string            `json:"ID"`
	Name           string            `json:"Name"`
	Image          string            `json:"Image"`
	TagCommit      string            `json:"TagCommit"`
	ImageCreatedAt string            `json:"ImageCreatedAt"`
	CreatedAt      string            `json:"CreatedAt"`
	Uptime         string            `json:"Uptime"`
	State          string            `json:"State"`
	Health         string            `json:"Health"`
	Run            bool              `json:"Run"`
	Restart        bool              `json:"Restart"`
	Labels         map[string]string `json:"Labels"`
	CPU            float64           `json:"CPU"`
	RAM            int64             `json:"RAM"`
	RAM_LIMIT      int64             `json:"RAM_LIMIT"`
}

type dockerAPIContainer struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	ImageID string            `json:"ImageID"`
	Created int64             `json:"Created"`
	State   string            `json:"State"`
	Status  string            `json:"Status"`
	Labels  map[string]string `json:"Labels"`
}

type dockerContainerInspect struct {
	Created string `json:"Created"`
	State   struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		StartedAt  string `json:"StartedAt"`
		FinishedAt string `json:"FinishedAt"`
		Health     *struct {
			Status string `json:"Status"`
		} `json:"Health"`
		Restarting   bool `json:"Restarting"`
		RestartCount int  `json:"RestartCount"`
	} `json:"State"`
	Config struct {
		Labels map[string]string `json:"Labels"`
		Image  string            `json:"Image"`
	} `json:"Config"`
}

type dockerImageInspect struct {
	Created  string   `json:"Created"`
	RepoTags []string `json:"RepoTags"`
}

// Структура для stats

type dockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage        uint64   `json:"total_usage"`
			PercpuUsage       []uint64 `json:"percpu_usage"`
			UsageInKernelmode uint64   `json:"usage_in_kernelmode"`
			UsageInUsermode   uint64   `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
}

func GetContainers() ([]Container, error) {
	debug := false
	if os.Getenv("DEBUG") == "true" {
		debug = true
	}
	log.Println("[docker-dashboard] GetContainers: start (net/http raw)")
	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
	}
	client := &http.Client{Transport: tr}
	url := "http://unix/containers/json?all=1"
	log.Printf("[docker-dashboard] GET %s", url)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("[docker-dashboard] http.Get error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("[docker-dashboard] status: %d", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[docker-dashboard] read body error: %v", err)
		return nil, err
	}
	if debug {
		log.Printf("[docker-dashboard] raw body: %s", string(body))
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("docker API status %d", resp.StatusCode)
	}
	var apiContainers []dockerAPIContainer
	if err := json.Unmarshal(body, &apiContainers); err != nil {
		log.Printf("[docker-dashboard] json.Unmarshal error: %v", err)
		return nil, err
	}
	log.Printf("[docker-dashboard] containers found: %d", len(apiContainers))
	var result []Container
	for _, c := range apiContainers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimLeft(c.Names[0], "/")
		}

		// Получаем подробную информацию о контейнере
		inspectURL := fmt.Sprintf("http://unix/containers/%s/json", c.ID)
		inspectResp, err := client.Get(inspectURL)
		if err != nil {
			log.Printf("[docker-dashboard] inspect error: %v", err)
			continue
		}
		inspectBody, err := ioutil.ReadAll(inspectResp.Body)
		inspectResp.Body.Close()
		if err != nil {
			log.Printf("[docker-dashboard] inspect read error: %v", err)
			continue
		}
		var inspect dockerContainerInspect
		if err := json.Unmarshal(inspectBody, &inspect); err != nil {
			log.Printf("[docker-dashboard] inspect unmarshal error: %v", err)
			continue
		}

		// Получаем информацию об образе
		imageCreatedAt := ""
		tagCommit := ""
		if c.ImageID != "" {
			imageURL := fmt.Sprintf("http://unix/images/%s/json", c.ImageID)
			imageResp, err := client.Get(imageURL)
			if err == nil {
				imageBody, err := ioutil.ReadAll(imageResp.Body)
				imageResp.Body.Close()
				if err == nil {
					var imageInfo dockerImageInspect
					if err := json.Unmarshal(imageBody, &imageInfo); err == nil {
						imageCreatedAt = imageInfo.Created
						if len(imageInfo.RepoTags) > 0 {
							tagCommit = imageInfo.RepoTags[0]
						}
					}
				}
			}
		}

		// create image/create container — ISO8601 (оригинал)
		createdAt := inspect.Created
		imageCreated := imageCreatedAt

		// uptime — человекочитаемый (29h26m2s)
		uptimeVal := ""
		if inspect.State.StartedAt != "" && inspect.State.Status == "running" {
			start, err := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)
			if err == nil {
				dur := time.Since(start)
				uptimeVal = dur.Truncate(time.Second).String()
			}
		}

		// restart — bool
		restart := false
		if inspect.State.RestartCount > 0 {
			restart = true
		}

		// tag|commit — из labels или RepoTags
		// Удаляю повторное объявление tagCommit, оставляю только присваивание
		if v, ok := c.Labels["org.quickex.frontend.commit"]; ok && v != "" {
			tagCommit = v
		} else if tagCommit == "" && len(tagCommit) == 0 && c.ImageID != "" {
			imageURL := fmt.Sprintf("http://unix/images/%s/json", c.ImageID)
			imageResp, err := client.Get(imageURL)
			if err == nil {
				imageBody, err := ioutil.ReadAll(imageResp.Body)
				imageResp.Body.Close()
				if err == nil {
					var imageInfo dockerImageInspect
					if err := json.Unmarshal(imageBody, &imageInfo); err == nil {
						if len(imageInfo.RepoTags) > 0 {
							tagCommit = imageInfo.RepoTags[0]
						}
					}
				}
			}
		}

		health := ""
		if inspect.State.Health != nil {
			health = inspect.State.Health.Status
		}
		shortID := c.ID
		if len(shortID) > 12 {
			shortID = shortID[:12]
		}
		// Форматируем временные метки — больше не нужно, оставляем ISO8601
		cpuPercent := 0.0
		ramUsage := int64(0)
		ramLimit := int64(0)
		// Получаем stats
		statsURL := fmt.Sprintf("http://unix/containers/%s/stats?stream=false", c.ID)
		statsResp, err := client.Get(statsURL)
		if err == nil {
			statsBody, err := io.ReadAll(statsResp.Body)
			statsResp.Body.Close()
			if err == nil {
				var stats dockerStats
				if err := json.Unmarshal(statsBody, &stats); err == nil {
					// CPU
					deltaCPU := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
					deltaSystem := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
					onlineCPUs := float64(stats.CPUStats.OnlineCPUs)
					if onlineCPUs == 0 {
						onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
					}
					if deltaSystem > 0 && onlineCPUs > 0 {
						cpuPercent = (deltaCPU / deltaSystem) * onlineCPUs * 100.0
					}
					// RAM
					ramUsage = int64(stats.MemoryStats.Usage)
					ramLimit = int64(stats.MemoryStats.Limit)
				}
			}
		}
		result = append(result, Container{
			ID:             shortID,
			Name:           name,
			Image:          c.Image,
			TagCommit:      tagCommit,
			ImageCreatedAt: imageCreated,
			CreatedAt:      createdAt,
			Uptime:         uptimeVal,
			State:          inspect.State.Status,
			Health:         health,
			Run:            inspect.State.Running,
			Restart:        restart,
			Labels:         c.Labels,
			CPU:            cpuPercent,
			RAM:            ramUsage,
			RAM_LIMIT:      ramLimit,
		})
	}
	return result, nil
}
