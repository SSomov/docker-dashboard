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

		let firstMessageReceived = false;
		let firstMessageResolver = null;
		const firstMessagePromise = new Promise((resolve, reject) => {
			firstMessageResolver = resolve;
			// Таймаут для первого сообщения (10 секунд)
			setTimeout(() => {
				if (!firstMessageReceived) {
					reject(new Error("Timeout waiting for first message"));
				}
			}, 10000);
		});

		ws.onopen = () => {
			console.log("WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (!firstMessageReceived) {
					firstMessageReceived = true;
					if (firstMessageResolver) {
						firstMessageResolver(data);
					}
				}
				if (callbacks.onMessage) callbacks.onMessage(data);
			} catch (error) {
				console.error("Error parsing WebSocket message:", error);
				if (!firstMessageReceived && firstMessageResolver) {
					firstMessageReceived = true;
					firstMessageResolver(null);
				}
			}
		};

		ws.onerror = (error) => {
			console.error("WebSocket error:", error);
			if (!firstMessageReceived && firstMessageResolver) {
				firstMessageReceived = true;
				firstMessageResolver(null);
			}
			if (callbacks.onError) callbacks.onError(error);
		};

		ws.onclose = () => {
			console.log("WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectContainersWebSocket(callbacks), 2000);
		};

		// Возвращаем промис для первого сообщения
		return firstMessagePromise;
	}

	function connectHostInfoWebSocket(callbacks) {
		const wsUrl = getWebSocketUrl("ws/hostinfo");

		wsHostInfo = new WebSocket(wsUrl);

		let firstMessageReceived = false;
		let firstMessageResolver = null;
		const firstMessagePromise = new Promise((resolve, reject) => {
			firstMessageResolver = resolve;
			// Таймаут для первого сообщения (10 секунд)
			setTimeout(() => {
				if (!firstMessageReceived) {
					reject(new Error("Timeout waiting for first message"));
				}
			}, 10000);
		});

		wsHostInfo.onopen = () => {
			console.log("HostInfo WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		wsHostInfo.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				if (!firstMessageReceived) {
					firstMessageReceived = true;
					if (firstMessageResolver) {
						firstMessageResolver(data);
					}
				}
				if (callbacks.onMessage) callbacks.onMessage(data);
			} catch (error) {
				console.error("Error parsing hostinfo WebSocket message:", error);
				if (!firstMessageReceived && firstMessageResolver) {
					firstMessageReceived = true;
					firstMessageResolver(null);
				}
			}
		};

		wsHostInfo.onerror = (error) => {
			console.error("HostInfo WebSocket error:", error);
			if (!firstMessageReceived && firstMessageResolver) {
				firstMessageReceived = true;
				firstMessageResolver(null);
			}
			if (callbacks.onError) callbacks.onError(error);
		};

		wsHostInfo.onclose = () => {
			console.log("HostInfo WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectHostInfoWebSocket(callbacks), 2000);
		};

		// Возвращаем промис для первого сообщения
		return firstMessagePromise;
	}

	function connectStatsWebSocket(callbacks) {
		const wsUrl = getWebSocketUrl("ws/containers/stats");

		wsStats = new WebSocket(wsUrl);

		let firstMessageReceived = false;
		let firstMessageResolver = null;
		const firstMessagePromise = new Promise((resolve, reject) => {
			firstMessageResolver = resolve;
			// Таймаут для первого сообщения (10 секунд)
			setTimeout(() => {
				if (!firstMessageReceived) {
					reject(new Error("Timeout waiting for first message"));
				}
			}, 10000);
		});

		wsStats.onopen = () => {
			console.log("Stats WebSocket connected");
			if (callbacks.onOpen) callbacks.onOpen();
		};

		wsStats.onmessage = (event) => {
			try {
				const stats = JSON.parse(event.data);
				if (!firstMessageReceived) {
					firstMessageReceived = true;
					if (firstMessageResolver) {
						firstMessageResolver(stats);
					}
				}
				if (callbacks.onMessage) callbacks.onMessage(stats);
			} catch (error) {
				console.error("Error parsing stats WebSocket message:", error);
				if (!firstMessageReceived && firstMessageResolver) {
					firstMessageReceived = true;
					firstMessageResolver(null);
				}
			}
		};

		wsStats.onerror = (error) => {
			console.error("Stats WebSocket error:", error);
			if (!firstMessageReceived && firstMessageResolver) {
				firstMessageReceived = true;
				firstMessageResolver(null);
			}
			if (callbacks.onError) callbacks.onError(error);
		};

		wsStats.onclose = () => {
			console.log("Stats WebSocket disconnected, reconnecting...");
			if (callbacks.onClose) callbacks.onClose();
			// Переподключение через 2 секунды
			setTimeout(() => connectStatsWebSocket(callbacks), 2000);
		};

		// Возвращаем промис для первого сообщения
		return firstMessagePromise;
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
