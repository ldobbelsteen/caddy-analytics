import "@fontsource/open-sans"
import "./style.sass"
import * as tools from "./tools.js"
import { Chart, LineElement, PointElement, LineController, CategoryScale, LinearScale, Tooltip } from "chart.js"
Chart.register(LineElement, PointElement, LineController, CategoryScale, LinearScale, Tooltip)

async function updateData() {

	// Fetch statistics/analytics
	const config = {
		cache: "no-store"
	}
	const raw = await fetch("/stats", config)
	const json = await raw.json()

	// Update general log statistics
	const header = document.querySelector("header")
	header.querySelector("#range").textContent = tools.unixToDate(json.firstStampUnix) + " to " + tools.unixToDate(json.lastStampUnix)
	header.querySelector("#duration").textContent = json.parseDurationSeconds.toFixed(2) + " seconds"
	header.querySelector("#size").textContent = tools.bytesToHumanReadable(json.logSizeBytes)
	header.querySelector("#directory").textContent = json.logDirectory

	// Sort all virtual hosts by total request count
	const menu = document.querySelector("nav")
	const counters = json.hostCounters
	const sortedHosts = Object.keys(counters).sort((a, b) => {
		const aVal = counters[a].totalMetrics.requests
		const bVal = counters[b].totalMetrics.requests
		return bVal - aVal
	})

	// Populate host selection menu with the hosts
	sortedHosts.forEach(host => {
		const button = document.createElement("div")
		button.textContent = host
		button.onclick = () => {
			recreateHourlyChart("general-stats-hourly", json, host)
		}
		menu.appendChild(button)
	})
}

updateData()

function recreateHourlyChart(container, data, host) {
	const div = document.getElementById(container)
	tools.removeAllChildren(div)
	const canvas = document.createElement("canvas")
	canvas.width = 600
	canvas.height = 400
	div.appendChild(canvas)

	const lastHour = tools.roundUnixDownToHour(data.lastStampUnix)
	const firstHour = lastHour - (4 * 24 * 60 * 60)

	data = data.hostCounters[host].hourlyMetrics
	const processedData = []
	let currentHour = firstHour
	do {
		const timeString = (new Date(currentHour * 1000)).toLocaleString()
		const requestCount = data[currentHour] != undefined ? data[currentHour].requests : 0
		processedData.push({
			x: timeString,
			y: requestCount
		})
		currentHour += 3600
	} while (currentHour <= lastHour)

	const chart = new Chart(canvas, {
		type: "line",
		data: {
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: processedData
			}]
		}
	})

	return chart
}
