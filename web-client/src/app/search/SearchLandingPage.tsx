'use client'

import React, { useState } from 'react'
import styles from './SearchLandingPage.module.css'
import { Button } from '@mui/base'
import { SourceDocument } from '@/app/search/types'
import { CitationBubble } from '@/app/search/CitationBubble'
import { SourceCard } from '@/app/search/SourceCard'
import { Card, CardContent, CircularProgress } from '@mui/material'
import CodeBlock from '@/app/CodeBlock'

type BaseChunk = {
    ID: string
}

type TextChunk = {
    type: 'text'
    value: string
}

type CodeBlockChunk = {
    type: 'codeblock'
    value: string
}

type CitationChunk = {
    type: 'citation'
    value: number
}

type AnswerChunk = BaseChunk & (TextChunk | CitationChunk | CodeBlockChunk)

const APIURL = process.env.NEXT_PUBLIC_API_URL

function toURL(query: string) {
    console.log(APIURL)
    return `${APIURL}/search?q=${encodeURIComponent(query)}&corpus=web`
}

export default function SearchLandingPage() {
    const [query, setQuery] = useState('')
    const [documents, setDocuments] = useState<SourceDocument[]>([])
    const [answerChunks, setAnswerChunks] = useState<AnswerChunk[]>([])
    const [hoveredCitationIndex, setHoveredCitationIndex] = useState<
        null | number
    >(null)
    const [isSearchResponseLoading, setIsSearchResponseLoading] = useState(true)

    const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setQuery(event.target.value)
    }

    const handleKeyPress = (event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'Enter') {
            handleSearch()
        }
    }

    const handleSearch = () => {
        setDocuments([])
        setAnswerChunks([])
        setIsSearchResponseLoading(true)

        const withSpinnerCanceller = (
            listener: (this: EventSource, event: MessageEvent<any>) => any
        ) => {
            function wrapped(this: EventSource, event: MessageEvent<any>) {
                setIsSearchResponseLoading(false)

                listener.call(this, event)
            }

            return wrapped
        }

        const eventSource = new EventSource(toURL(query), {
            withCredentials: true,
        })

        eventSource.addEventListener(
            'text',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)

                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'text', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'citation',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)

                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'citation', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'documentsreference',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setDocuments(data)
            })
        )

        eventSource.addEventListener(
            'codeblock',
            withSpinnerCanceller((event) => {
                const data = JSON.parse(event.data)
                setAnswerChunks((prev) => [
                    ...prev,
                    { type: 'codeblock', value: data, ID: data.ID },
                ])
            })
        )

        eventSource.addEventListener(
            'done',
            withSpinnerCanceller((event) => {
                eventSource.close()
            })
        )

        eventSource.onerror = (err) => {
            console.error(err)
            eventSource.close()
            setIsSearchResponseLoading(false)
        }
    }

    return (
        <div className={styles.searchRoot}>
            <div className={styles.searchBar}>
                <input
                    placeholder="Why are peppers spicy?"
                    value={query}
                    onChange={handleInputChange}
                    onKeyDown={handleKeyPress}
                    className={styles.searchInput}
                />
                <Button onClick={handleSearch} className={styles.searchButton}>
                    Search
                </Button>
            </div>
            {isSearchResponseLoading && (
                <CircularProgress
                    style={{ color: 'var(--brand-teal)' }}
                    className={styles.loadingSpinner}
                />
            )}
            {documents.length > 0 && (
                <div className={styles.resultSection}>
                    <h2>Sources</h2>
                    <div className={styles.sourceContainer}>
                        {documents.map((document, index) => {
                            return (
                                <SourceCard
                                    source={document.webReference!}
                                    isHoveredViaCitation={
                                        hoveredCitationIndex === index
                                    }
                                    key={document.webReference?.link}
                                />
                            )
                        })}
                    </div>
                </div>
            )}
            {answerChunks.length > 0 && (
                <div className={styles.resultSection}>
                    <h2>Results</h2>
                    <Card className={styles.answerCard}>
                        <CardContent className={styles.cardContent}>
                            {answerChunks.map((ac) => (
                                <AnswerChunk
                                    ac={ac}
                                    setHoveredCitationIndex={
                                        setHoveredCitationIndex
                                    }
                                    documents={documents}
                                    key={ac.ID}
                                />
                            ))}
                        </CardContent>
                    </Card>
                </div>
            )}
        </div>
    )
}

interface AnswerChunkProps {
    ac: AnswerChunk
    documents: SourceDocument[]
    setHoveredCitationIndex: React.Dispatch<React.SetStateAction<number | null>>
}

function AnswerChunk({
    ac,
    documents,
    setHoveredCitationIndex,
}: AnswerChunkProps) {
    const handleCitationClick = (citationIndex: number) => {
        const link = documents[citationIndex]?.webReference?.link
        if (!link) return
        window.open(link, '_blank', 'noopener noreferrer')
    }

    switch (ac.type) {
        case 'text':
            return <span className={styles.answerText}>{ac.value}</span>
        case 'citation':
            return (
                <CitationBubble
                    onClick={() => handleCitationClick(ac.value)}
                    label={ac.value + 1}
                    onMouseEnter={() => setHoveredCitationIndex(ac.value)}
                    onMouseLeave={() => setHoveredCitationIndex(null)}
                />
            )
        case 'codeblock':
            const lines = ac.value.split('\n')
            const language = lines[0].match(/```(\w+)/)?.[1] || ''
            const code = lines.slice(1, -1).join('\n')
            return <CodeBlock language={language} code={code} />
        default:
            return <span>Unsupported answer chunk</span>
    }
}
