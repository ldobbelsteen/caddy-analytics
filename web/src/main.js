import * as tools from "./tools.js"
import {
	Chart,
	ArcElement,
	DoughnutController,
	BarElement,
	BarController,
	LineElement,
	PointElement,
	LineController,
	CategoryScale,
	LinearScale,
	Tooltip
} from "chart.js"

Chart.register(
	ArcElement,
	DoughnutController,
	BarElement,
	BarController,
	LineElement,
	PointElement,
	LineController,
	CategoryScale,
	LinearScale,
	Tooltip
)

window.onload = async () => {

	// Fetch statistics/analytics
	const raw = await fetch("/data", { cache: "no-store" })
	const json = await raw.json()

	// Write general log statistics
	const header = document.querySelector("header")
	header.querySelector("#range").textContent = tools.unixToDate(json.firstStampUnix) + " to " + tools.unixToDate(json.lastStampUnix)
	header.querySelector("#duration").textContent = json.parseDurationSeconds.toFixed(2) + " seconds"
	header.querySelector("#size").textContent = tools.bytesToHumanReadable(json.logSizeBytes)
	header.querySelector("#directory").textContent = json.logDirectory

	// Sort hosts by their total request count and populate the menu
	const menu = document.querySelector("nav")
	tools.removeAllChildren(menu)
	Object.keys(json.hostCounters).sort((a, b) => {
		const aVal = json.hostCounters[a].totalMetrics.requests
		const bVal = json.hostCounters[b].totalMetrics.requests
		return bVal - aVal
	}).forEach(host => {
		const button = document.createElement("div")
		button.textContent = host
		button.onclick = () => {
			const main = document.querySelector("main")
			tools.removeAllChildren(main)
			createHourlyChart(main, json, host)
			createOperatingSystemsChart(main, json, host)
			createBrowsersChart(main, json, host)
			createDevicesChart(main, json, host)
			createLanguageChart(main, json, host)
		}
		menu.appendChild(button)
	})
}

function createHourlyChart(container, data, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const end = tools.roundUnixDownToHour(data.lastStampUnix)
	let current = end - (4 * 24 * 60 * 60)
	const filtered = []
	do {
		const timeString = (new Date(current * 1000)).toLocaleString()
		const metrics = data.hostCounters[host].hourlyMetrics
		if (metrics[current] != undefined) {
			filtered.push({
				x: timeString,
				y: metrics[current].nonBotVisitors
			})
		} else {
			filtered.push({
				x: timeString,
				y: 0
			})
		}
		current += 60 * 60
	} while (current <= end)
	
	new Chart(canvas, {
		type: "line",
		data: {
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: filtered
			}]
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true
				}
			}
		}
	})
}

function createOperatingSystemsChart(container, data, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const filtered = []
	const scopedData = data.hostCounters[host].nonBotVisitorSystems
	Object.keys(scopedData).forEach(key => {
		filtered.push({
			x: key,
			y: scopedData[key]
		})
	})
	filtered.sort((a, b) => {
		return b.y - a.y
	})

	new Chart(canvas, {
		type: "bar",
		data: {
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: filtered
			}]
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true
				}
			}
		}
	})
}

function createDevicesChart(container, fullData, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const scopedData = fullData.hostCounters[host].nonBotVisitorDevices
	const labels = Object.keys(scopedData)
	const data = Object.values(scopedData)

	new Chart(canvas, {
		type: "bar",
		data: {
			labels: labels,
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: data
			}]
		},
		options: {
			maintainAspectRatio: false
		}
	})
}

function createBrowsersChart(container, fullData, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const scopedData = fullData.hostCounters[host].nonBotVisitorBrowsers
	const labels = Object.keys(scopedData)
	const data = Object.values(scopedData)

	new Chart(canvas, {
		type: "bar",
		data: {
			labels: labels,
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: data
			}]
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true
				}
			}
		}
	})
}

function createLanguageChart(container, fullData, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const scopedData = fullData.hostCounters[host].nonBotVisitorPrefLanguages
	const labels = Object.keys(scopedData)
	const data = Object.values(scopedData)

	new Chart(canvas, {
		type: "bar",
		data: {
			labels: labels,
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: data
			}]
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true
				}
			}
		}
	})
}
