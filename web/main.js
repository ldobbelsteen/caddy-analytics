import './style.css'
import '@fontsource/open-sans'

const statsURL = "stats"

async function init() {
	const paragraph = document.createElement('h1')
	paragraph.textContent = "Caddy Analytics"
	document.body.appendChild(paragraph)

	const data = await fetch(statsURL, {cache: "no-store"})
	const stats = await data.json()
	const string = JSON.stringify(stats, null, 4)
	const text = document.createElement('pre')
	text.innerHTML = string
	document.body.appendChild(text)
}

init()
