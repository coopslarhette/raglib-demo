'use client'

import React, { useEffect } from 'react'
import dynamic from 'next/dynamic'
import SearchBar from './SearchBar'
import { CircularProgress, Tooltip } from '@mui/material'
import { useAnswerStream } from '@/app/search/use-answer-stream'
import styles from './SearchContainer.module.css'

const SearchResults = dynamic(() => import('./SearchResults'), { ssr: false })

interface SearchContainerProps {
    initialQuery: string
}

export default function SearchContainer({
    initialQuery,
}: SearchContainerProps) {
    const { answerChunks, documents, isResponseLoading, handleSearch } =
        useAnswerStream(initialQuery)

    useEffect(() => {
        if (initialQuery.length > 0) {
            handleSearch(initialQuery)
        }
    }, [initialQuery])

    return (
        <div className={styles.root}>
            <SearchBar initialQuery={initialQuery} onSearch={handleSearch} />
            {isResponseLoading ? (
                <Tooltip title="Sorry, sometimes the backend has a cold start (free hosting).">
                    <CircularProgress style={{ color: 'var(--brand-teal)' }} />
                </Tooltip>
            ) : (
                <SearchResults
                    documents={documents}
                    answerChunks={answerChunks}
                />
            )}
        </div>
    )
}
