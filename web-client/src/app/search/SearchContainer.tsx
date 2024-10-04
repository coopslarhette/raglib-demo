'use client'

import React, { useEffect } from 'react'
import dynamic from 'next/dynamic'
import SearchBar from './SearchBar'
import { CircularProgress } from '@mui/material'
import { useAnswerStream } from '@/app/search/use-answer-stream'

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
        <>
            <SearchBar initialQuery={initialQuery} onSearch={handleSearch} />
            {isResponseLoading ? (
                <CircularProgress style={{ color: 'var(--brand-teal)' }} />
            ) : (
                <SearchResults
                    documents={documents}
                    answerChunks={answerChunks}
                />
            )}
        </>
    )
}
