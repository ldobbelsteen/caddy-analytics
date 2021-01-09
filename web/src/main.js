import "@fontsource/open-sans"
import "./style.sass"
import * as tools from "./tools.js"

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
		button.classList.add("button")
		button.textContent = host
		button.onclick = () => {
			const data = json["counters"][button.textContent]
			const pre = document.querySelector("#json")
			pre.textContent = JSON.stringify(data, undefined, 4)
		}
		menu.appendChild(button)
	})
}

updateData()
