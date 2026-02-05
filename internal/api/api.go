package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"docker-dashboard/internal/containers"
	"docker-dashboard/internal/hostinfo"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключения с любого origin
	},
}

func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/containers", getContainersHandler).Methods("GET")
	// WebSocket эндпоинт для контейнеров
	r.HandleFunc("/ws/containers", containersWebSocketHandler)
	// Новый эндпоинт для системных метрик
	r.HandleFunc("/api/hostinfo", getHostInfoHandler).Methods("GET")
	// WebSocket эндпоинт для hostinfo
	r.HandleFunc("/ws/hostinfo", hostinfoWebSocketHandler)
	// WebSocket эндпоинт для логов контейнера
	r.HandleFunc("/ws/containers/{id}/logs", containerLogsWebSocketHandler)
	// WebSocket эндпоинт для перезагрузки контейнера
	r.HandleFunc("/ws/containers/{id}/restart", containerRestartWebSocketHandler)
}

type containerGroup struct {
	ProjectName string                 `json:"project_name,omitempty"`
	Containers  []containers.Container `json:"containers"`
}

type containersResponse struct {
	SnapshotTime    time.Time              `json:"snapshot_time"`
	Total           int                    `json:"total"`
	Containers      []containers.Container `json:"containers"` // Для обратной совместимости
	Groups          []containerGroup       `json:"groups"`
	LogsShow        bool                   `json:"logs_show"`
	ContainerRestart bool                  `json:"container_restart"`
}

func groupContainers(containerList []containers.Container) []containerGroup {
	groupsMap := make(map[string][]containers.Container)
	var ungrouped []containers.Container

	for _, container := range containerList {
		if container.ComposeProject != "" {
			groupsMap[container.ComposeProject] = append(groupsMap[container.ComposeProject], container)
		} else {
			ungrouped = append(ungrouped, container)
		}
	}

	// Сортируем контейнеры внутри каждой группы по имени
	for projectName := range groupsMap {
		sort.Slice(groupsMap[projectName], func(i, j int) bool {
			return groupsMap[projectName][i].Name < groupsMap[projectName][j].Name
		})
	}

	// Сортируем контейнеры без группы по имени
	if len(ungrouped) > 0 {
		sort.Slice(ungrouped, func(i, j int) bool {
			return ungrouped[i].Name < ungrouped[j].Name
		})
	}

	// Собираем группы и сортируем их по имени проекта
	var groups []containerGroup
	for projectName, containers := range groupsMap {
		groups = append(groups, containerGroup{
			ProjectName: projectName,
			Containers:  containers,
		})
	}

	// Сортируем группы по имени проекта (пустые имена идут в конец)
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].ProjectName == "" {
			return false
		}
		if groups[j].ProjectName == "" {
			return true
		}
		return groups[i].ProjectName < groups[j].ProjectName
	})

	// Добавляем группу без проекта в конец
	if len(ungrouped) > 0 {
		groups = append(groups, containerGroup{
			ProjectName: "",
			Containers:  ungrouped,
		})
	}

	return groups
}

func getLogsShow() bool {
	logsShow := os.Getenv("LOGS_SHOW")
	if logsShow == "" {
		return false
	}
	value, err := strconv.ParseBool(logsShow)
	if err != nil {
		return false
	}
	return value
}

func getContainerRestart() bool {
	containerRestart := os.Getenv("CONTAINER_RESTART")
	if containerRestart == "" {
		return false
	}
	value, err := strconv.ParseBool(containerRestart)
	if err != nil {
		return false
	}
	return value
}

func getContainersHandler(w http.ResponseWriter, r *http.Request) {
	containerList, err := containers.GetContainers()
	if err != nil {
		http.Error(w, "Failed to get containers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := containersResponse{
		SnapshotTime:    time.Now(),
		Total:           len(containerList),
		Containers:      containerList,
		Groups:          groupContainers(containerList),
		LogsShow:        getLogsShow(),
		ContainerRestart: getContainerRestart(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func containersWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Канал для обработки закрытия соединения клиентом
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Отправляем данные сразу при подключении
	sendContainersData(conn)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := sendContainersData(conn); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func sendContainersData(conn *websocket.Conn) error {
	containerList, err := containers.GetContainers()
	if err != nil {
		log.Printf("Failed to get containers: %v", err)
		return err
	}

	response := containersResponse{
		SnapshotTime:    time.Now(),
		Total:           len(containerList),
		Containers:      containerList,
		Groups:          groupContainers(containerList),
		LogsShow:        getLogsShow(),
		ContainerRestart: getContainerRestart(),
	}

	return conn.WriteJSON(response)
}

func getHostInfoHandler(w http.ResponseWriter, r *http.Request) {
	metrics, err := hostinfo.GetSystemMetrics()
	if err != nil {
		http.Error(w, "Failed to get system metrics", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func hostinfoWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Канал для обработки закрытия соединения клиентом
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Отправляем данные сразу при подключении
	sendHostInfoData(conn)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := sendHostInfoData(conn); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func sendHostInfoData(conn *websocket.Conn) error {
	metrics, err := hostinfo.GetSystemMetrics()
	if err != nil {
		log.Printf("Failed to get hostinfo: %v", err)
		return err
	}

	return conn.WriteJSON(metrics)
}

func containerLogsWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	containerID := vars["id"]
	if containerID == "" {
		http.Error(w, "Container ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Создаем HTTP клиент для Docker API с закрытием соединений
	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
	}
	defer tr.CloseIdleConnections()

	client := &http.Client{
		Transport: tr,
		Timeout:   0, // Без таймаута для streaming
	}

	// Канал для обработки закрытия соединения клиентом
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return
			}
		}
	}()

	// Запрашиваем логи с follow=true для получения потока
	// Используем timestamps=false для упрощения, но все равно нужно обработать заголовки
	logsURL := "http://unix/containers/" + containerID + "/logs?follow=true&stdout=true&stderr=true&tail=100&timestamps=false"
	log.Printf("[docker-dashboard] GET %s", logsURL)

	req, err := http.NewRequest("GET", logsURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		conn.WriteJSON(map[string]string{"error": "Failed to create request"})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to get logs: %v", err)
		conn.WriteJSON(map[string]string{"error": "Failed to get logs: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Docker API returned status %d", resp.StatusCode)
		conn.WriteJSON(map[string]string{"error": "Docker API returned status " + string(rune(resp.StatusCode))})
		return
	}

	// Docker API возвращает логи в формате: [8 байт заголовка][данные]
	// Заголовок: [stream type (1 байт)][padding (3 байта)][размер данных (4 байта, big-endian)]
	// Stream type: 1 = stdout, 2 = stderr
	const maxLogSize = 64 * 1024 // Максимальный размер одной строки лога (64KB)
	for {
		select {
		case <-done:
			return
		default:
		}

		// Читаем заголовок (8 байт)
		header := make([]byte, 8)
		n, err := resp.Body.Read(header)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading header: %v", err)
			break
		}
		if n < 8 {
			break
		}

		// Извлекаем размер данных из заголовка (байты 4-7, big-endian)
		size := int(header[4])<<24 | int(header[5])<<16 | int(header[6])<<8 | int(header[7])
		if size <= 0 {
			continue
		}
		// Ограничиваем размер для предотвращения утечек памяти
		if size > maxLogSize {
			log.Printf("Log line too large (%d bytes), truncating to %d", size, maxLogSize)
			size = maxLogSize
		}

		// Читаем данные
		data := make([]byte, size)
		read := 0
		for read < size {
			n, err := resp.Body.Read(data[read:])
			if err != nil && err != io.EOF {
				log.Printf("Error reading log data: %v", err)
				return
			}
			if n == 0 {
				break
			}
			read += n
		}

		// Отправляем через WebSocket
		logLine := string(data[:read])
		if err := conn.WriteJSON(map[string]string{"log": logLine}); err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
	}
}

func containerRestartWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем что функция включена
	if !getContainerRestart() {
		http.Error(w, "Container restart is disabled", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	containerID := vars["id"]
	if containerID == "" {
		http.Error(w, "Container ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Создаем HTTP клиент для Docker API
	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", "/var/run/docker.sock")
		},
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
	}
	defer tr.CloseIdleConnections()

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	// Выполняем POST запрос к Docker API для перезагрузки контейнера
	restartURL := "http://unix/containers/" + containerID + "/restart"
	log.Printf("[docker-dashboard] POST %s", restartURL)

	req, err := http.NewRequest("POST", restartURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		conn.WriteJSON(map[string]string{"status": "error", "message": "Failed to create request: " + err.Error()})
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to restart container: %v", err)
		conn.WriteJSON(map[string]string{"status": "error", "message": "Failed to restart container: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// Docker API возвращает 204 No Content при успешном перезапуске
	if resp.StatusCode == http.StatusNoContent {
		conn.WriteJSON(map[string]string{"status": "success", "message": "Container restarted successfully"})
		return
	}

	// Читаем тело ответа для получения деталей ошибки
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read error response: %v", err)
		conn.WriteJSON(map[string]string{"status": "error", "message": "Docker API returned status " + strconv.Itoa(resp.StatusCode)})
		return
	}

	log.Printf("Docker API returned status %d: %s", resp.StatusCode, string(body))
	conn.WriteJSON(map[string]string{"status": "error", "message": "Failed to restart container: " + string(body)})
}
