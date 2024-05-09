export type Passage = {
    text: string
}

export type Source = 'web' | 'personal'

export type WebReference = {
    title: string
    link: string
    displayedLink: string
    snippet: string
    date: string
    author: string
    favicon: string
    thumbnail: string
}

export type SourceDocument = {
    passages: Passage[]
    source: Source
    webReference?: WebReference
}
