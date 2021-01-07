import "@fontsource/open-sans"
import "./style.sass"
import * as tools from "./tools.js"

function updateData() {
	fetch("/stats", {
		cache: "no-store"
	}).then(response => response.json())
		.then(json => {
			const header = document.querySelector("#header")
			header.querySelector("#range").textContent = tools.unixToDate(json.firstStampUnix) + " to " + tools.unixToDate(json.lastStampUnix)
			header.querySelector("#duration").textContent = json.parseDurationSeconds + " seconds"
			header.querySelector("#size").textContent = tools.bytesToHumanReadable(json.logSizeBytes)
			header.querySelector("#directory").textContent = json.logDirectory

			const menu = document.querySelector("#menu")
			for (const host in json["counters"]) {
				const button = document.createElement("div")
				button.textContent = host
				button.onclick = () => {
					const data = json["counters"][button.textContent]
					const pre = document.querySelector("#json")
					pre.textContent = JSON.stringify(data, undefined, 4)
				}
				menu.appendChild(button)
			}
		})
}

updateData()

/*
function createHttpBarChart(data, container) {
	data = Object.keys(data).map(key => ({ key: String(key), value: data[key]}))
	container = d3.select(container)
	const margin = {
		top: 20,
		bottom: 20,
		left: 30,
		right: 20
	}
	const width = 600
	const height = 400
	const chartWidth = width - margin.left - margin.right
	const chartHeight = height - margin.top - margin.bottom

	const svg = container.append("svg")
		.attr("width", width)
		.attr("height", height)
	const chart = svg.append("g")
		.attr("transform", `translate(${margin.left}, ${margin.top})`)

	const yScale = d3.scaleLinear()
		.range([chartHeight, 0])
		.domain([0, d3.max(data, d => +d.value)])
	chart.append("g")
		.call(d3.axisLeft(yScale))	
	const xScale = d3.scaleBand()
		.range([0, chartWidth])
		.domain(data.map(d => d.key))
		.padding(0.1)
	chart.append("g")
		.attr("transform", `translate(0, ${chartHeight})`)
		.call(d3.axisBottom(xScale))
	
	chart.selectAll()
		.data(data)
		.enter().append("rect")
			.attr("x", d => xScale(d.key))
			.attr("y", d => yScale(d.value))
			.attr("height", d => chartHeight - yScale(d.value))
			.attr("width", xScale.bandwidth())
}
*/
