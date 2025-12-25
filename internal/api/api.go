package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
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
}

type containerGroup struct {
	ProjectName string                `json:"project_name,omitempty"`
	Containers  []containers.Container `json:"containers"`
}

type containersResponse struct {
	SnapshotTime time.Time         `json:"snapshot_time"`
	Total        int               `json:"total"`
	Containers   []containers.Container `json:"containers"` // Для обратной совместимости
	Groups       []containerGroup  `json:"groups"`
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

func getContainersHandler(w http.ResponseWriter, r *http.Request) {
	containerList, err := containers.GetContainers()
	if err != nil {
		http.Error(w, "Failed to get containers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := containersResponse{
		SnapshotTime: time.Now(),
		Total:        len(containerList),
		Containers:   containerList,
		Groups:       groupContainers(containerList),
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

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Отправляем данные сразу при подключении
	sendContainersData(conn)

	for {
		select {
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
		SnapshotTime: time.Now(),
		Total:        len(containerList),
		Containers:   containerList,
		Groups:       groupContainers(containerList),
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

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Отправляем данные сразу при подключении
	sendHostInfoData(conn)

	for {
		select {
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
