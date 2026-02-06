package containers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Глобальный HTTP клиент для Docker API для переиспользования соединений
var (
	dockerClient     *http.Client
	dockerClientOnce sync.Once
)

func getDockerClient() *http.Client {
	dockerClientOnce.Do(func() {
		tr := &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
		dockerClient = &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		}
	})
	return dockerClient
}

type DeployResources struct {
	CPULimit          string `json:"CPULimit,omitempty"`
	MemoryLimit       string `json:"MemoryLimit,omitempty"`
	CPUReservation    string `json:"CPUReservation,omitempty"`
	MemoryReservation string `json:"MemoryReservation,omitempty"`
}

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
	ComposeProject string            `json:"ComposeProject,omitempty"`
	DeployResources *DeployResources `json:"DeployResources,omitempty"`
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
	HostConfig struct {
		Memory            int64 `json:"Memory"`
		MemoryReservation int64 `json:"MemoryReservation"`
		CpuQuota          int64 `json:"CpuQuota"`
		CpuPeriod         int64 `json:"CpuPeriod"`
		NanoCpus          int64 `json:"NanoCpus"`
	} `json:"HostConfig"`
}

type dockerImageInspect struct {
	Created  string   `json:"Created"`
	RepoTags []string `json:"RepoTags"`
}

type dockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64   `json:"total_usage"`
			Percpu     []uint64 `json:"percpu_usage,omitempty"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs      uint32 `json:"online_cpus"`
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

type ContainerStats struct {
	ID          string  `json:"ID"`
	CPUUsage    float64 `json:"CPUUsage"`    // CPU usage in cores
	MemoryUsage int64   `json:"MemoryUsage"` // Memory usage in bytes
}

func filterLabels(labels map[string]string) map[string]string {
	if labels == nil {
		return nil
	}

	labelPrefix := os.Getenv("LABEL_PREFIX")
	labelPrefixExclude := os.Getenv("LABEL_PREFIX_EXCLUDE")

	// Если указан LABEL_PREFIX, показываем только labels с этим префиксом
	if labelPrefix != "" {
		filtered := make(map[string]string)
		for key, value := range labels {
			if strings.HasPrefix(key, labelPrefix) {
				filtered[key] = value
			}
		}
		return filtered
	}

	// Если указан LABEL_PREFIX_EXCLUDE, показываем все кроме labels с этим префиксом
	if labelPrefixExclude != "" {
		filtered := make(map[string]string)
		for key, value := range labels {
			if !strings.HasPrefix(key, labelPrefixExclude) {
				filtered[key] = value
			}
		}
		return filtered
	}

	// Если ничего не указано, возвращаем все labels
	return labels
}

func formatMemory(bytes int64) string {
	if bytes == 0 {
		return ""
	}
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	if bytes >= GB {
		return fmt.Sprintf("%.1fG", float64(bytes)/float64(GB))
	}
	if bytes >= MB {
		return fmt.Sprintf("%.1fM", float64(bytes)/float64(MB))
	}
	if bytes >= KB {
		return fmt.Sprintf("%.1fK", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%dB", bytes)
}

func formatCPU(quota, period, nanoCpus int64) string {
	if nanoCpus > 0 {
		// NanoCpus представляет CPU в наносекундах (1 CPU = 1e9 nanoseconds)
		cpus := float64(nanoCpus) / 1e9
		return fmt.Sprintf("%.1f", cpus)
	}
	if quota > 0 && period > 0 {
		// CpuQuota / CpuPeriod дает количество CPU
		cpus := float64(quota) / float64(period)
		return fmt.Sprintf("%.1f", cpus)
	}
	return ""
}

func parseResources(inspect dockerContainerInspect) *DeployResources {
	resources := &DeployResources{}

	// Parse CPU limit
	resources.CPULimit = formatCPU(
		inspect.HostConfig.CpuQuota,
		inspect.HostConfig.CpuPeriod,
		inspect.HostConfig.NanoCpus,
	)

	// Parse Memory limit
	resources.MemoryLimit = formatMemory(inspect.HostConfig.Memory)

	// Parse CPU reservation (same logic as limit, but typically not set separately in HostConfig)
	// For now, we'll leave it empty unless there's a specific field
	resources.CPUReservation = ""

	// Parse Memory reservation
	resources.MemoryReservation = formatMemory(inspect.HostConfig.MemoryReservation)

	// Return nil if no resources are set
	if resources.CPULimit == "" && resources.MemoryLimit == "" &&
		resources.CPUReservation == "" && resources.MemoryReservation == "" {
		return nil
	}

	return resources
}

func GetContainers() ([]Container, error) {
	debug := false
	if os.Getenv("DEBUG") == "true" {
		debug = true
	}
	log.Println("[docker-dashboard] GetContainers: start (net/http raw)")
	client := getDockerClient()
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

	type containerResult struct {
		container Container
		index     int
		err       error
	}

	resultChan := make(chan containerResult, len(apiContainers))
	var wg sync.WaitGroup

	for i, c := range apiContainers {
		wg.Add(1)
		go func(idx int, container dockerAPIContainer) {
			defer wg.Done()

			name := ""
			if len(container.Names) > 0 {
				name = strings.TrimLeft(container.Names[0], "/")
			}

			// Получаем подробную информацию о контейнере
			inspectURL := fmt.Sprintf("http://unix/containers/%s/json", container.ID)
			inspectResp, err := client.Get(inspectURL)
			if err != nil {
				log.Printf("[docker-dashboard] inspect error: %v", err)
				resultChan <- containerResult{err: err, index: idx}
				return
			}
			inspectBody, err := io.ReadAll(inspectResp.Body)
			inspectResp.Body.Close()
			if err != nil {
				log.Printf("[docker-dashboard] inspect read error: %v", err)
				resultChan <- containerResult{err: err, index: idx}
				return
			}
			var inspect dockerContainerInspect
			if err := json.Unmarshal(inspectBody, &inspect); err != nil {
				log.Printf("[docker-dashboard] inspect unmarshal error: %v", err)
				resultChan <- containerResult{err: err, index: idx}
				return
			}

			// Получаем информацию об образе (только один раз)
			imageCreatedAt := ""
			tagCommit := ""
			if v, ok := container.Labels["org.quickex.frontend.commit"]; ok && v != "" {
				tagCommit = v
			}

			if container.ImageID != "" && tagCommit == "" {
				imageURL := fmt.Sprintf("http://unix/images/%s/json", container.ImageID)
				imageResp, err := client.Get(imageURL)
				if err == nil {
					imageBody, err := io.ReadAll(imageResp.Body)
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

			health := ""
			if inspect.State.Health != nil {
				health = inspect.State.Health.Status
			}
			shortID := container.ID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			filteredLabels := filterLabels(container.Labels)

			// Извлекаем compose project до фильтрации меток
			composeProject := ""
			if container.Labels != nil {
				if project, ok := container.Labels["com.docker.compose.project"]; ok {
					composeProject = project
				}
			}

			// Парсим ресурсы deploy
			deployResources := parseResources(inspect)

			resultChan <- containerResult{
				container: Container{
					ID:             shortID,
					Name:           name,
					Image:          container.Image,
					TagCommit:      tagCommit,
					ImageCreatedAt: imageCreated,
					CreatedAt:      createdAt,
					Uptime:         uptimeVal,
					State:          inspect.State.Status,
					Health:         health,
					Run:            inspect.State.Running,
					Restart:        restart,
					Labels:         filteredLabels,
					ComposeProject: composeProject,
					DeployResources: deployResources,
				},
				index: idx,
			}
		}(i, c)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Собираем результаты с сохранением порядка
	results := make([]*Container, len(apiContainers))
	for res := range resultChan {
		if res.err == nil {
			results[res.index] = &res.container
		}
	}

	// Фильтруем nil значения и формируем финальный результат
	result := make([]Container, 0, len(apiContainers))
	for _, res := range results {
		if res != nil {
			result = append(result, *res)
		}
	}

	return result, nil
}

// GetContainerStats получает статистику использования CPU и RAM для контейнера
func GetContainerStats(containerID string) (*ContainerStats, error) {
	client := getDockerClient()
	statsURL := fmt.Sprintf("http://unix/containers/%s/stats?stream=false&one-shot=true", containerID)
	
	resp, err := client.Get(statsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("docker stats API returned status %d", resp.StatusCode)
	}

	var stats dockerStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("failed to decode stats: %w", err)
	}

	// Вычисляем CPU usage в ядрах
	var cpuCores float64
	
	// Проверяем, что есть данные для вычисления дельты
	if stats.PreCPUStats.SystemCPUUsage > 0 && stats.CPUStats.SystemCPUUsage > stats.PreCPUStats.SystemCPUUsage {
		cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
		
		if systemDelta > 0 && stats.CPUStats.OnlineCPUs > 0 {
			// Вычисляем процент использования CPU
			cpuPercent := (cpuDelta / systemDelta) * float64(stats.CPUStats.OnlineCPUs) * 100.0
			// Конвертируем процент в количество ядер
			cpuCores = cpuPercent / 100.0
			// Ограничиваем количеством доступных ядер
			if cpuCores > float64(stats.CPUStats.OnlineCPUs) {
				cpuCores = float64(stats.CPUStats.OnlineCPUs)
			}
			// Не показываем отрицательные значения
			if cpuCores < 0 {
				cpuCores = 0
			}
		}
	}

	// Получаем использование памяти
	memoryUsage := int64(stats.MemoryStats.Usage)

	return &ContainerStats{
		ID:          containerID,
		CPUUsage:    cpuCores,
		MemoryUsage: memoryUsage,
	}, nil
}

// GetContainersStats получает статистику для всех запущенных контейнеров
func GetContainersStats() ([]ContainerStats, error) {
	client := getDockerClient()
	url := "http://unix/containers/json?all=1"
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get containers list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("docker API status %d", resp.StatusCode)
	}

	var apiContainers []dockerAPIContainer
	if err := json.NewDecoder(resp.Body).Decode(&apiContainers); err != nil {
		return nil, fmt.Errorf("failed to decode containers: %w", err)
	}

	type statsResult struct {
		stats *ContainerStats
		err   error
	}

	statsChan := make(chan statsResult, len(apiContainers))
	var wg sync.WaitGroup

	for _, apiContainer := range apiContainers {
		// Получаем статистику только для запущенных контейнеров
		if apiContainer.State != "running" {
			continue
		}

		wg.Add(1)
		go func(containerID string) {
			defer wg.Done()
			// Используем полный ID для получения статистики
			stats, err := GetContainerStats(containerID)
			if err != nil {
				// Игнорируем ошибки для остановленных контейнеров
				if strings.Contains(err.Error(), "No such container") ||
					strings.Contains(err.Error(), "is not running") {
					return
				}
				log.Printf("[docker-dashboard] Failed to get stats for container %s: %v", containerID, err)
				return
			}
			// Используем короткий ID для совместимости с фронтендом
			if len(containerID) > 12 {
				stats.ID = containerID[:12]
			} else {
				stats.ID = containerID
			}
			statsChan <- statsResult{stats: stats, err: nil}
		}(apiContainer.ID)
	}

	go func() {
		wg.Wait()
		close(statsChan)
	}()

	var result []ContainerStats
	for res := range statsChan {
		if res.stats != nil {
			result = append(result, *res.stats)
		}
	}

	return result, nil
}
