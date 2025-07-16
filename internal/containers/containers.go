package containers

import (
	// "fmt"
	"os"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

type Container struct {
	Name           string
	ID             string
	Image          string
	TagCommit      string
	ImageCreatedAt string
	CreatedAt      string
	Uptime         string
	State          string
	Health         string
	Run            bool
	Restart        bool
	// RAMUsage      string
	// CPUUsage      string
	// ContainerSize string
	Labels map[string]string
}

func GetContainers() ([]Container, error) {
	prefix := os.Getenv("LABEL_PREFIX")
	if prefix == "" {
		prefix = "org.example"
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return nil, err
	}

	var result []Container
	for _, c := range containers {
		containerInfo, err := client.InspectContainerWithOptions(docker.InspectContainerOptions{
			ID: c.ID,
		})
		if err != nil {
			return nil, err
		}
		imageInfo, err := client.InspectImage(c.Image)
		if err != nil {
			return nil, err
		}

		// statsChan := make(chan *docker.Stats)
		// done := make(chan bool)
		// go func() {
		// 	client.Stats(docker.StatsOptions{
		// 		ID:     c.ID,
		// 		Stats:  statsChan,
		// 		Stream: false,
		// 		Done:   done,
		// 	})
		// }()

		// stats := <-statsChan
		// close(done)

		// ramUsage := fmt.Sprintf("%.2f MiB", float64(stats.MemoryStats.Usage)/1024/1024)
		// cpuUsage := fmt.Sprintf("%.2f%%", calculateCPUPercent(stats))
		// containerSize := fmt.Sprintf("%v MiB", containerInfo.SizeRootFs/1024/1024)

		container := Container{
			Name:           strings.TrimLeft(c.Names[0], "/"),
			ID:             c.ID[:12],
			Image:          getImageName(c.Image),
			TagCommit:      getTagCommit(c.Image),
			ImageCreatedAt: imageInfo.Created.Format(time.RFC3339),
			CreatedAt:      containerInfo.Created.Format(time.RFC3339),
			Uptime:         calculateUptime(containerInfo.Created),
			State:          containerInfo.State.String(),
			Health:         containerInfo.State.Health.Status,
			Run:            containerInfo.State.Running,
			Restart:        containerInfo.State.Restarting,
			// RAMUsage:      ramUsage,
			// CPUUsage:      cpuUsage,
			// ContainerSize: containerSize,
			Labels: filterLabels(imageInfo.Config.Labels, prefix),
		}
		result = append(result, container)
	}

	return result, nil
}

func filterLabels(labels map[string]string, prefix string) map[string]string {
	filtered := make(map[string]string)
	for key, value := range labels {
		if strings.HasPrefix(key, prefix) {
			filtered[key] = value
		}
	}
	return filtered
}

func getImageName(fullName string) string {
	parts := strings.Split(fullName, ":")
	return parts[0]
}

func getTagCommit(fullName string) string {
	parts := strings.Split(fullName, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// func calculateCPUPercent(stats *docker.Stats) float64 {
// 	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
// 	systemDelta := float64(stats.CPUStats.SystemCPUUsage - stats.PreCPUStats.SystemCPUUsage)
// 	numberCPUs := float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
// 	if systemDelta > 0.0 && cpuDelta > 0.0 {
// 		return (cpuDelta / systemDelta) * numberCPUs * 100.0
// 	}
// 	return 0.0
// }

func calculateUptime(createdAt time.Time) string {
	uptime := time.Since(createdAt)
	return uptime.Round(time.Second).String()
}
