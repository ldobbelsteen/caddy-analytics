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
	const counters = json["counters"]
	const sortedHosts = Object.keys(counters).sort((a, b) => {
		const aVal = counters[a]["total"]["requests"]
		const bVal = counters[b]["total"]["requests"]
		return bVal - aVal
	})

	// Populate host selection menu with the hosts
	sortedHosts.forEach(host => {
		const button = document.createElement("div")
		button.textContent = host
		button.onclick = () => {
			const data = json["counters"][button.textContent]
			renderHourly(data["hourly"], "general-stats-hourly")
		}
		menu.appendChild(button)
	})
}

updateData()

function renderHourly(rawData, canvasId) {
	const canvas = document.getElementById(canvasId)
	const lastHour = Object.keys(rawData).reduce((lastHour, hour) => parseInt(hour) > parseInt(lastHour) ? parseInt(hour) : parseInt(lastHour))
	const firstHour = lastHour - (6 * 24 * 60 * 60)

	const data = []
	let currentHour = firstHour
	do {
		data.push({
			x: (new Date(currentHour * 1000)).toLocaleString(),
			y: (rawData[currentHour] ?? 0).requests ?? 0
		})
		currentHour += 3600
	} while (currentHour <= lastHour)

	new Chart(canvas, {
		type: "line",
		data: {
			datasets: [{
				backgroundColor: "rgb(255, 99, 132)",
				borderColor: "rgb(255, 99, 132)",
				data: data
			}]
		}
	})
}
