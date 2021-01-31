import Chart from "chart.js"
import { parallelSort } from "./tools.js"

Chart.defaults.global.legend.display = false
Chart.defaults.global.maintainAspectRatio = false
Chart.defaults.global.elements.rectangle.backgroundColor = "#bdd6f7"

export default [
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.systems
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.bots
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.browsers
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.countries
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.devices
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.encodings
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.visitors.languages
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.requests.ciphers
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.requests.contents
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.requests.cryptos
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.requests.methods
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
	canvas => {
		const chart = new Chart(canvas, {
			type: "bar",
			data: {
				labels: [],
				datasets: [{
					data: [],
				}],
			},
		})
		return host => {
			const data = host.requests.protocols
			const [values, labels] = parallelSort([Object.values(data), Object.keys(data)])
			chart.config.data.labels = labels
			chart.config.data.datasets[0].data = values
			chart.update()
		}
	},
]
