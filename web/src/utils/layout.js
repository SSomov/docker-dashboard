export function checkMobile() {
	return window.matchMedia("(max-width: 768px)").matches;
}

export function updateFixedHeights() {
	const headerEl = document.querySelector("header");
	const metricsBarEl = document.querySelector(".metrics-bar");
	const filterBarEl = document.querySelector(".filter-bar");
	const groupTagsBarEl = document.querySelector(".group-tags-bar");

	const headerHeight = headerEl ? headerEl.offsetHeight : 0;
	const metricsBarHeight = metricsBarEl ? metricsBarEl.offsetHeight : 0;
	const filterBarHeight = filterBarEl ? filterBarEl.offsetHeight : 0;
	const groupTagsBarHeight = groupTagsBarEl ? groupTagsBarEl.offsetHeight : 0;
	const totalFixedHeight =
		headerHeight + metricsBarHeight + filterBarHeight + groupTagsBarHeight;

	return {
		headerHeight,
		metricsBarHeight,
		filterBarHeight,
		groupTagsBarHeight,
		totalFixedHeight,
	};
}
