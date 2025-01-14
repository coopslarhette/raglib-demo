'use client'

import React, { useMemo, useState } from 'react'
import styles from './SearchResults.module.css'
import { SourceDocument, AnswerChunk } from './types'
import { SourceCard } from './SourceCard'
import { Card, CardContent, CircularProgress } from '@mui/material'
import { MarkdownRenderer } from '@/app/search/MarkdownRenderer'

interface SearchResultsProps {
    documents: SourceDocument[]
    answerChunks: AnswerChunk[]
}

export default function SearchResults({
    documents,
    answerChunks,
}: SearchResultsProps) {
    const [hoveredCitationIndex, setHoveredCitationIndex] = useState<
        null | number
    >(null)

    // This is a bit of a hack for now, could def refactor things to not need this
    const content = useMemo(() => {
        return answerChunks.reduce((acc, chunk: AnswerChunk) => {
            if (chunk.type === 'citation') {
                return acc + `<cited>${chunk.value}</cited>`
            }
            return acc + chunk.value
        }, '')
    }, [answerChunks])

    return (
        <>
            {documents.length > 0 && (
                <div className={styles.resultSection}>
                    <h2 className={styles.headers}>Sources</h2>
                    <div className={styles.sourceContainer}>
                        {documents.map((document, index) => (
                            <SourceCard
                                source={document.webReference!}
                                isHoveredViaCitation={
                                    hoveredCitationIndex === index
                                }
                                key={document.webReference?.link}
                            />
                        ))}
                    </div>
                </div>
            )}
            {content.length > 0 && (
                <div className={styles.resultSection}>
                    <h2 className={styles.headers}>Synthesis</h2>
                    <Card className={styles.answerCard}>
                        <CardContent className={styles.cardContent}>
                            <MarkdownRenderer content={content} setHoveredCitationIndex={setHoveredCitationIndex} />
                        </CardContent>
                    </Card>
                </div>
            )}
        </>
    )
}
