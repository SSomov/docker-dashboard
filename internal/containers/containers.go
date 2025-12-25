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
	"sync"
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
	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}
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
			inspectBody, err := ioutil.ReadAll(inspectResp.Body)
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

			health := ""
			if inspect.State.Health != nil {
				health = inspect.State.Health.Status
			}
			shortID := container.ID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			filteredLabels := filterLabels(container.Labels)

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
