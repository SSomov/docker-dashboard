export function createWebSocketStore() {
	let ws = null;
	let wsHostInfo = null;
	let wsStats = null;

	function getWebSocketUrl(path) {
		const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
		return `${protocol}//${window.location.host}${window.location.pathname}${path}`;
	}

	function connectContainersWebSocket(callbacks) {
		const wsUrl = getWebSocketUrl("ws/containers");

		ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			console.log("WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (callbacks.onMessage) callbacks.onMessage(data);
			} catch (error) {
				console.error("Error parsing WebSocket message:", error);
			}
		};

		ws.onerror = (error) => {
			console.error("WebSocket error:", error);
			if (callbacks.onError) callbacks.onError(error);
		};

		ws.onclose = () => {
			console.log("WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectContainersWebSocket(callbacks), 2000);
		};
	}

	function connectHostInfoWebSocket(callbacks) {
		const wsUrl = getWebSocketUrl("ws/hostinfo");

		wsHostInfo = new WebSocket(wsUrl);

		wsHostInfo.onopen = () => {
			console.log("HostInfo WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		wsHostInfo.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (callbacks.onMessage) callbacks.onMessage(data);
			} catch (error) {
				console.error("Error parsing hostinfo WebSocket message:", error);
			}
		};

		wsHostInfo.onerror = (error) => {
			console.error("HostInfo WebSocket error:", error);
			if (callbacks.onError) callbacks.onError(error);
		};

		wsHostInfo.onclose = () => {
			console.log("HostInfo WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectHostInfoWebSocket(callbacks), 2000);
		};
	}

	function connectStatsWebSocket(callbacks) {
		const wsUrl = getWebSocketUrl("ws/containers/stats");

		wsStats = new WebSocket(wsUrl);

		wsStats.onopen = () => {
			console.log("Stats WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		wsStats.onmessage = (event) => {
			try {
				const stats = JSON.parse(event.data);
				if (callbacks.onMessage) callbacks.onMessage(stats);
			} catch (error) {
				console.error("Error parsing stats WebSocket message:", error);
			}
		};

		wsStats.onerror = (error) => {
			console.error("Stats WebSocket error:", error);
			if (callbacks.onError) callbacks.onError(error);
		};

		wsStats.onclose = () => {
			console.log("Stats WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectStatsWebSocket(callbacks), 2000);
		};
	}

	function connectRestartWebSocket(containerId, callbacks) {
		const wsUrl = getWebSocketUrl(`ws/containers/${containerId}/restart`);

		const wsRestart = new WebSocket(wsUrl);

		wsRestart.onopen = () => {
			console.log("Restart WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		wsRestart.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (callbacks.onMessage) callbacks.onMessage(data);
			} catch (error) {
				console.error("Error parsing restart response:", error);
				if (callbacks.onError) callbacks.onError(error);
			}
			wsRestart.close();
		};

		wsRestart.onerror = (error) => {
			console.error("WebSocket error for restart:", error);
			if (callbacks.onError) callbacks.onError(error);
			wsRestart.close();
		};

		wsRestart.onclose = () => {
			console.log("Restart WebSocket disconnected");
			if (callbacks.onClose) callbacks.onClose();
		};
	}

	function disconnectAll() {
		if (ws) {
			ws.close();
			ws = null;
		}
		if (wsHostInfo) {
			wsHostInfo.close();
			wsHostInfo = null;
		}
		if (wsStats) {
			wsStats.close();
			wsStats = null;
		}
	}

	return {
		connectContainersWebSocket,
		connectHostInfoWebSocket,
		connectStatsWebSocket,
		connectRestartWebSocket,
		disconnectAll,
	};
}
