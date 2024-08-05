export function dateToHumanReadable(dateString: string): string {
    const date = new Date(dateString);

    if (isNaN(date.getTime())) {
        return "";
    }

    const options: Intl.DateTimeFormatOptions = {
        year: 'numeric',
        month: 'long',
        day: 'numeric'
    };

    return date.toLocaleDateString('en-US', options);
}
