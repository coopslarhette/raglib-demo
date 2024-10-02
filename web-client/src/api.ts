const APIURL = process.env.NEXT_PUBLIC_API_URL

export function toSearchURL(query: string) {
    return `${APIURL}/search?q=${encodeURIComponent(query)}&corpus=web`
}
