// Remove all children of an HTML element
export function removeAllChildren(element) {
	while (element.firstChild) {
		element.removeChild(element.firstChild)
	}
}

// Perform sort on an array of arrays based on the first array
export function parallelSort(arrays) {
	const sortableArray = arrays[0]
	const indices = Object.keys(sortableArray)
	indices.sort((a, b) => sortableArray[b] - sortableArray[a])
	const sortByIndices = (arr, ind) => ind.map(i => arr[i])
	return arrays.map(arr => sortByIndices(arr, indices))
}

// Convert a number of bytes to a more human-readable format
export function bytesToHumanReadable(bytes) {
	const base = 1024
	if (bytes <= 0) {
		return "0 B"
	} else {
		let order = Math.floor(Math.log(bytes) / Math.log(base))
		const suffixes = ["B", "kB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"]
		if (order > suffixes.length - 1) {
			order = suffixes.length - 1
		}
		const value = (bytes / Math.pow(base, order)).toFixed(2)
		return value + " " + suffixes[order]
	}
}

// Convert a Unix timestamp to date string
export function unixToDate(unix) {
	const milliseconds = unix * 1000
	const date = new Date(milliseconds)
	const string = date.toLocaleDateString()
	return string
}
