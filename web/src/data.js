import { bytesToHumanReadable, unixToDate } from "./tools.js"

// Download and format the data
export default async () => {
	const raw = await fetch("/data", { cache: "no-store" })
	const data = await raw.json()

	data.firstStamp = unixToDate(data.firstStamp)
	data.lastStamp = unixToDate(data.lastStamp)
	data.parseDuration = data.parseDuration.toFixed(2) + " seconds"
	data.sizeBytes = bytesToHumanReadable(data.sizeBytes)

	return data
}
