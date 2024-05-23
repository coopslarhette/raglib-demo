'use client'
import React, { useState } from 'react'
import styles from './SearchLandingPage.module.css'
import { Button } from '@mui/base'
import { SourceDocument } from '@/app/search/types'
import { CitationBubble } from '@/app/search/CitationBubble'
import { SourceCard } from '@/app/search/SourceCard'
import { setHttpClientAndAgentOptions } from 'next/dist/server/setup-http-agent-env'
import { number } from 'prop-types'

type TextChunk = {
    type: 'text'
    value: string
}

type CitationChunk = {
    type: 'citation'
    value: number
}

type AnswerChunk = TextChunk | CitationChunk

export default function SearchLandingPage() {
    const [query, setQuery] = useState('')
    const [documents, setDocuments] = useState<SourceDocument[]>([])
    const [answerChunks, setAnswerChunks] = useState<AnswerChunk[]>([])
    const [hoveredCitationIndex, setHoveredCitationIndex] = useState<
        null | number
    >(null)

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
        const url = `http://localhost:5000/search?q=${encodeURIComponent(query)}&corpus=web`

        const eventSource = new EventSource(url, {
            withCredentials: true,
        })

        eventSource.addEventListener('text', (event) => {
            const data = JSON.parse(event.data)

            setAnswerChunks((prev) => [...prev, { type: 'text', value: data }])
        })

        eventSource.addEventListener('citation', (event) => {
            const data = JSON.parse(event.data)

            setAnswerChunks((prev) => [
                ...prev,
                { type: 'citation', value: data },
            ])
        })

        eventSource.addEventListener('documentsreference', (event) => {
            const data = JSON.parse(event.data)
            setDocuments(data)
        })

        eventSource.addEventListener('done', (event) => {
            eventSource.close()
        })

        eventSource.onerror = (err) => {
            console.error(err)
            eventSource.close()
        }
    }

    const handleCitationClick = (citationIndex: number) => {
        const link = documents[citationIndex]?.webReference?.link
        if (!link) return
        window.open(link, '_blank', 'noopener noreferrer')
    }

    return (
        <div className={styles.searchRoot}>
            <div className={styles.searchBar}>
                <input
                    placeholder="Search..."
                    value={query}
                    onChange={handleInputChange}
                    onKeyDown={handleKeyPress}
                    className={styles.searchInput}
                />
                <Button onClick={handleSearch} className={styles.searchButton}>
                    Search
                </Button>
            </div>
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
                                />
                            )
                        })}
                    </div>
                </div>
            )}
            {answerChunks.length > 0 && (
                <div className={styles.resultSection}>
                    <h2>Results</h2>
                    <div>
                        {answerChunks.map((ac) => {
                            return ac.type === 'text' ? (
                                <span>{ac.value}</span>
                            ) : (
                                <CitationBubble
                                    onClick={() => handleCitationClick(ac.value)}
                                    label={ac.value + 1}
                                    onMouseEnter={() =>
                                        setHoveredCitationIndex(ac.value)
                                    }
                                    onMouseLeave={() =>
                                        setHoveredCitationIndex(null)
                                    }
                                />
                            )
                        })}
                    </div>
                </div>
            )}
        </div>
    )
}
