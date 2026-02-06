export function formatTime(dateString) {
	if (!dateString) return "";
	try {
		const date = new Date(dateString);
		if (isNaN(date.getTime())) {
			console.warn("Invalid date:", dateString);
			return "";
		}
		return date.toLocaleTimeString("ru-RU", {
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
		});
	} catch (error) {
		console.error("Error formatting time:", error, dateString);
		return "";
	}
}

export function getCpuPercent(cpu) {
	return cpu?.length ? cpu[0] : 0;
}

export function getRamPercent(used, total) {
	return used && total ? (used / total) * 100 : 0;
}

export function getSwapPercent(swapTotal, swapFree) {
	return swapTotal > 0 ? ((swapTotal - swapFree) / swapTotal) * 100 : 0;
}
