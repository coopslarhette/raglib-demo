'use client'

import React, { useState } from 'react'
import styles from './SearchResults.module.css'
import { SourceDocument, AnswerChunk } from './types'
import { SourceCard } from './SourceCard'
import { Card, CardContent, CircularProgress } from '@mui/material'
import AnswerSection from './AnswerSection'

interface SearchResultsProps {
    documents: SourceDocument[]
    answerChunks: AnswerChunk[]
    isLoading: boolean
}

export default function SearchResults({
    documents,
    answerChunks,
    isLoading,
}: SearchResultsProps) {
    const [hoveredCitationIndex, setHoveredCitationIndex] = useState<
        null | number
    >(null)

    if (isLoading) {
        return (
            <CircularProgress
                style={{ color: 'var(--brand-teal)' }}
                className={styles.loadingSpinner}
            />
        )
    }

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
            {answerChunks.length > 0 && (
                <div className={styles.resultSection}>
                    <h2 className={styles.headers}>Synthesis</h2>
                    <Card className={styles.answerCard}>
                        <CardContent className={styles.cardContent}>
                            {answerChunks.map((ac) => (
                                <AnswerSection
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
        </>
    )
}
