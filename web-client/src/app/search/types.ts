export type Passage = {
    text: string
}

export type Corpus = 'web' | 'personal'

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
    corpus: Corpus
    webReference?: WebReference
}

export type ChunkType =
    | 'text'
    | 'citation'
    | 'documentsreference'
    | 'codeblock'
    | 'done'

export type BaseChunk = {
    ID: string
    type: ChunkType
}

export type TextChunk = {
    type: 'text'
    value: string
}

export type CodeBlockChunk = {
    type: 'codeblock'
    value: string
}

export type CitationChunk = {
    type: 'citation'
    value: number
}

export type AnswerChunk = BaseChunk &
    (TextChunk | CitationChunk | CodeBlockChunk)
