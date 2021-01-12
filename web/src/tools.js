// Convert a number of bytes to a more human-readable format
export function bytesToHumanReadable(bytes) {
	const base = 1024
	if (bytes <= 0) {
		return "0 B"
	} else {
		var order = Math.floor(Math.log(bytes) / Math.log(base))
		const suffixes = ["B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"]
		if (order > suffixes.length - 1) {
			order = suffixes.length - 1
		}
		const value = (bytes / Math.pow(base, order)).toFixed(2)
		return value + " " + suffixes[order]
	}
}

// Convert a unix timestamp to locale date string
export function unixToDate(unix) {
	const milliseconds = unix * 1000
	const date = new Date(milliseconds)
	const string = date.toLocaleDateString()
	return string
}

// Round down a unix timestamp down to the hour
export function roundUnixDownToHour(unix) {
	return unix - (unix % (60 * 60))
}

// Remove all children of an HTML element
export function removeAllChildren(element) {
	while (element.firstChild) {
		element.removeChild(element.firstChild)
	}
}
