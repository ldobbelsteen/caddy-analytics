import { removeAllChildren } from "./tools.js"
import Charts from "./charts.js"
import Data from "./data.js"

window.onload = async () => {
	const data = await Data()
	const header = document.querySelector("header")
	const menu = document.querySelector("nav")
	const main = document.querySelector("main")

	header.querySelector("#range").textContent = data.firstStamp + " to " + data.lastStamp
	header.querySelector("#duration").textContent = data.parseDuration
	header.querySelector("#size").textContent = data.sizeBytes
	header.querySelector("#directory").textContent = data.directory
	
	const dataUpdaters = []
	Charts.forEach(chart => {
		const div = document.createElement("div")
		main.append(div)
		const canvas = document.createElement("canvas")
		div.append(canvas)
		const updater = chart(canvas)
		dataUpdaters.push(updater)
	})
	
	const hosts = Object.keys(data.hosts)
	hosts.sort((a, b) => {
		return data.hosts[b].total.visitors - data.hosts[a].total.visitors
	})

	removeAllChildren(menu)
	hosts.forEach(host => {
		const button = document.createElement("div")
		button.textContent = host
		button.onclick = () => {
			dataUpdaters.forEach(updater => {
				updater(data.hosts[host])
			})
		}
		menu.appendChild(button)
	})
}
/*
function createHourlyChart(container, data, host) {
	const div = document.createElement("div")
	const canvas = document.createElement("canvas")
	container.appendChild(div)
	div.appendChild(canvas)

	const end = roundUnixDownToHour(data.lastStampUnix)
	let current = end - (4 * 24 * 60 * 60)
	const filtered = []
	do {
		const timeString = (new Date(current * 1000)).toLocaleString()
		const metrics = data.hostCounters[host].hourlyMetrics
		if (metrics[current] != undefined) {
			filtered.push({
				x: timeString,
				y: metrics[current].nonBotVisitors,
			})
		} else {
			filtered.push({
				x: timeString,
				y: 0,
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
				data: filtered,
			}],
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true,
				},
			},
		},
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
			y: scopedData[key],
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
				data: filtered,
			}],
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true,
				},
			},
		},
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
				data: data,
			}],
		},
		options: {
			maintainAspectRatio: false,
		},
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
				data: data,
			}],
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true,
				},
			},
		},
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
				data: data,
			}],
		},
		options: {
			maintainAspectRatio: false,
			scales: {
				y: {
					beginAtZero: true,
				},
			},
		},
	})
}
*/
