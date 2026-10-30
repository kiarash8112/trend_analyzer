/** "https://www.example.com/a" -> "example.com"; returns the input when it is not a URL. */
export function hostOf(url: string): string {
  try {
    return new URL(url).hostname.replace(/^www\./, '');
  } catch {
    return url;
  }
}

/** The first clause of a launch channel: "X (Twitter), the author's post" -> "X". */
export function shortChannel(channel: string): string {
  return channel.split(/[,(;.]/)[0].trim();
}

export function truncate(text: string, max: number): string {
  return text.length > max ? text.slice(0, max - 1).trimEnd() + '…' : text;
}
